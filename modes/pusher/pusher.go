package pusher

import (
	"bufio"
	"fmt"
	"github.com/alomerry/go-pusher/component/oss"
	"github.com/alomerry/go-pusher/share"
	"github.com/alomerry/go-pusher/utils"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type Pusher struct {
	localDirPath     string
	ossPrefix        string            // 待推送目录对应 oss 的前缀
	pushTimeout      uint32            // 推送超时时间 单位：秒
	hashCacheMapper  map[string]string // 上次缓存的文件 md5
	resultHashMapper map[string]string // 本次遍历的文件 md5
	providers        []string

	status string
	lock   *sync.RWMutex
}

func (p *Pusher) InitConfig() {
	p.localDirPath = cast.ToString(viper.GetStringMap("pusher")["local-directory"])
	p.ossPrefix = cast.ToString(viper.GetStringMap("pusher")["oss-object-prefix"])
	p.pushTimeout = cast.ToUint32(viper.GetStringMap("pusher")["push-timeout"])

	// 获取或创建 .oss_pusher_hash 文件，将文件列表记录到 map 中 TODO 文件数量过多可能会 oom
	p.hashCacheMapper = getCacheHashMap(p.localDirPath)
	p.resultHashMapper = make(map[string]string)

	p.providers = cast.ToStringSlice(viper.GetStringMap("pusher")["oss-provider"])

	p.lock = &sync.RWMutex{}
	p.status = share.TASK_STATUS_PENDING
}

func (p *Pusher) Run(ctx context.Context) error {
	defer p.done()

	if p.pushTimeout > 0 {
		ctx, _ = context.WithTimeout(ctx, time.Duration(p.pushTimeout)*time.Second)
	}

	// 遍历配置中文件夹下的文件，根据文件完整的相对路径l对比 .oss_pusher_hash 中的 md5，如果有新增或变动的文件，则推送到 oss
	err := fs.WalkDir(os.DirFS(p.localDirPath), ".", p.upsertFile)
	if err != nil {
		panic(err)
	}

	// 遍历完毕后，如果 hashCacheMapper 非空，则需要检查 oss 有没有对应 key 的文件，如果存在，需要删除（TODO）
	err = p.tryDelNotExistsFile()
	if err != nil {
		panic(err)
	}

	// 写入最新版的文件路径对应的 md5 映射到 .oss_pusher_hash 中去
	err = p.writeCache(p.localDirPath)
	if err != nil {
		panic(err)
	}
	return nil
}

func (p *Pusher) Done() bool {
	var done bool
	p.lock.RLock()
	done = p.status == share.TASK_STATUS_DONE
	p.lock.RUnlock()
	return done
}

func (p *Pusher) done() {
	p.setStatus(share.TASK_STATUS_DONE)
}

func (p *Pusher) process() {
	p.setStatus(share.TASK_STATUS_PROCESSING)
}

func (p *Pusher) setStatus(status string) {
	p.lock.Lock()
	p.status = status
	p.lock.Unlock()
}

func getCacheHashMap(cachePath string) map[string]string {
	hashCacheMapper := make(map[string]string)
	cacheFile := fmt.Sprintf("%s/%s", cachePath, share.CACHE_FILE_NAME)
	_, err := os.Stat(cacheFile)
	if err != nil {
		// TODO 不存在则创建
		if os.IsNotExist(err) {
			_, err = os.Create(cacheFile)
		}
		if err != nil {
			panic(".oss_pusher_hash not exists or create failed.")
		}
	} else {
		file, err := os.OpenFile(cacheFile, os.O_RDONLY, 0)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		reader := bufio.NewReader(file)
		for {
			line, isPrefix, err := reader.ReadLine()
			if len(line) > 0 && err != nil {
				panic(err)
			}
			if isPrefix {
				panic("file too large")
			}
			if err != nil {
				if err != io.EOF {
					panic(err)
				}
				break
			}
			infos := strings.Split(string(line), "[@]")
			hashCacheMapper[infos[0]] = infos[1]
		}
	}
	return hashCacheMapper
}

func (p *Pusher) upsertFile(relatePath string, d fs.DirEntry, err error) error {
	if err != nil {
		panic(err)
	}
	if d.IsDir() {
		return nil
	}
	if needIgnore(relatePath) {
		return nil
	}
	hash, err := utils.FileMD5(fmt.Sprintf("%s/%s", p.localDirPath, relatePath))
	if err != nil {
		panic(err)
	}

	localFilePath := fmt.Sprintf("%s/%s", p.localDirPath, relatePath)
	ossFilePath := fmt.Sprintf("%s/%s", p.ossPrefix, relatePath)
	cacheHash, exist := p.hashCacheMapper[relatePath]
	// oss 不存在或文件有变动时 upsert 到 oss
	if !exist || cacheHash != hash {
		key, err := oss.Client.Push(localFilePath, ossFilePath)
		if err != nil {
			panic(err)
		}

		fmt.Printf("推送文件 [%v] 到 oss [%v:%v]\n", localFilePath, ossFilePath, key)
	}

	p.resultHashMapper[relatePath] = hash
	delete(p.hashCacheMapper, relatePath)
	return nil
}

func (p *Pusher) tryDelNotExistsFile() error {
	needDelNotExists := viper.GetBool(cast.ToString(viper.GetStringMap("pusher")["oss-delete-not-exists"]))
	for key := range p.hashCacheMapper {
		if needDelNotExists {
			// ossFilePath := fmt.Sprintf("%s/%s", p.ossPrefix, key)
			// oss.Client.Delete(ossFilePath)
		}

		fmt.Printf("删除文件[%v]到 oss\n", key)
	}
	return nil
}

func needIgnore(filepath string) bool {
	if filepath == ".oss_pusher_hash" {
		return true
	}

	if strings.HasSuffix(filepath, ".toml") {
		return true
	}

	if strings.HasSuffix(filepath, ".DS_Store") {
		return true
	}
	return false
}

func (p *Pusher) writeCache(dirPath string) error {
	resultBuilder := &strings.Builder{}
	// 按文件名排序写入，防止文件频繁变更
	keys := make([]string, 0, len(p.resultHashMapper))
	for key := range p.resultHashMapper {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	for i := range keys {
		resultBuilder.WriteString(fmt.Sprintf("%s[@]%s\n", keys[i], p.resultHashMapper[keys[i]]))
	}

	cacheFile := fmt.Sprintf("%s/%s", dirPath, share.CACHE_FILE_NAME)
	err := os.WriteFile(cacheFile, []byte(resultBuilder.String()), os.FileMode(0777))
	if err != nil {
		return err
	}
	return nil
}
