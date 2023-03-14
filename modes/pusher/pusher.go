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
)

type Pusher struct {
	localDirPath     string
	ossPrefix        string
	hashCacheMapper  map[string]string
	resultHashMapper map[string]string
}

// Run
// - 新建 .oss_pusher_hash 记录文件列表和文件 md5 值
// - 获取文件列表 fileList
// - 读取 .oss_pusher_hash 中记录的 md5List
// - 遍历 fileList。如果 md5List 中不存在，则推送到 oss（如何回写到 markdown url 中，按照目录名？），并计算文件名及 md5  值到变量 x 中，移出 md5List 中该 key；如果 md5List 中存在并且值一致，记录到变量 x 中；如果值不一致，同不存在逻辑一致
// - 遍历完毕后，如果 md5List 非空，则需要检查 oss 有没有对应 key 的文件，如果存在，需要删除
// - 将变量 x 覆盖写入 oss_pusher_hash 中
func (p Pusher) Run(ctx context.Context) error {
	ctx = context.WithValue(ctx, "status", "running")
	err := fs.WalkDir(os.DirFS(p.localDirPath), ".", p.upsertFile)
	if err != nil {
		panic(err)
	}

	err = p.tryDelNotExistsFile()
	if err != nil {
		panic(err)
	}

	err = p.writeCache(p.localDirPath)
	if err != nil {
		panic(err)
	}
	return nil
}

func (p Pusher) InitConfig() {
	p.localDirPath = cast.ToString(viper.GetStringMap("pusher")["local-directory"])
	p.ossPrefix = cast.ToString(viper.GetStringMap("pusher")["oss-object-prefix"])
	p.hashCacheMapper = getCacheHashMap(p.localDirPath)
}

func getCacheHashMap(cachePath string) map[string]string {
	hashCacheMapper := make(map[string]string)
	cacheFile := fmt.Sprintf("%s/%s", cachePath, share.CACHE_FILE_NAME)
	_, err := os.Stat(cacheFile)
	if err != nil && os.IsNotExist(err) {
		panic(".oss_pusher_hash not exists")
	} else {
		file, err := os.OpenFile(cacheFile, os.O_RDONLY, 0)
		if err != nil {
			panic(err)
		}

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

func (p Pusher) upsertFile(relatePath string, d fs.DirEntry, err error) error {
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
	if !exist {
		_, err := oss.Client.Push(localFilePath, ossFilePath)
		if err != nil {
			panic(err)
		}

		fmt.Printf("推送新文件[%v]到 oss\n", relatePath)
	} else {
		if cacheHash != hash {
			_, err := oss.Client.Push(localFilePath, ossFilePath)
			if err != nil {
				panic(err)
			}
			fmt.Printf("更新文件[%v]到 oss\n", relatePath)
		}
	}

	resultHashMapper := make(map[string]string)
	resultHashMapper[relatePath] = hash
	delete(p.hashCacheMapper, relatePath)
	return nil
}

func (p Pusher) tryDelNotExistsFile() error {
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

func (p Pusher) writeCache(dirPath string) error {
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

func (p Pusher) Done(ctx context.Context) bool {
	return cast.ToBool(ctx.Value("status"))
}
