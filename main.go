package main

import (
	"bufio"
	"fmt"
	"github.com/alomerry/OSSPusher/pusher"
	"github.com/alomerry/OSSPusher/share"
	"github.com/alomerry/OSSPusher/utils"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
)

// - 新建 .oss_pusher_hash 记录文件列表和文件 md5 值
// - 获取文件列表 fileList
// - 读取 .oss_pusher_hash 中记录的 md5List
// - 遍历 fileList。如果 md5List 中不存在，则推送到 oss（如何回写到 markdown url 中，按照目录名？），并计算文件名及 md5  值到变量 x 中，移出 md5List 中该 key；如果 md5List 中存在并且值一致，记录到变量 x 中；如果值不一致，同不存在逻辑一致
// - 遍历完毕后，如果 md5List 非空，则需要检查 oss 有没有对应 key 的文件，如果存在，需要删除
// - 将变量 x 覆盖写入 oss_pusher_hash 中

var (
	configPath       = pflag.String("configPath", "", "path of configuration")
	hashCacheMapper  map[string]string
	resultHashMapper map[string]string
	dirPath          string
	ossPrefix        string

	cpuProfile = "cpu.profile" // write cpu profile
	memProfile = "mem.profile" // write memory profile
)

func init() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	share.ExPath = filepath.Dir(ex)
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

func loadConfig() {
	viper.BindPFlag("configPath", pflag.Lookup("configPath"))
	pflag.Parse()

	viper.SetConfigFile(fmt.Sprintf("%s/%s", share.ExPath, strings.TrimPrefix(strings.TrimPrefix(viper.GetString("configPath"), share.ExPath), "/")))
	err := viper.MergeInConfig()
	if err != nil {
		panic(err)
	}
	dirPath = viper.GetString("local-directory")
	ossPrefix = viper.GetString("oss-object-prefix")
}

func pprofProfile() func() {
	return func() {
		if strings.HasSuffix(cpuProfile, ".profile") {
			f, err := os.Create(fmt.Sprintf("%s/pprof/%s", share.ExPath, cpuProfile))
			if err != nil {
				panic(err)
			}
			defer f.Close()
			if err := pprof.StartCPUProfile(f); err != nil {
				panic(err)
			}
			defer pprof.StopCPUProfile()
		}

		if strings.HasSuffix(memProfile, ".profile") {
			defer func() {
				f, err := os.Create(fmt.Sprintf("%s/pprof/%s", share.ExPath, memProfile))
				if err != nil {
					panic(err)
				}
				defer f.Close()
				runtime.GC() // get up-to-date statistics
				if err := pprof.WriteHeapProfile(f); err != nil {
					panic(err)
				}
			}()
		}
	}
}

func main() {
	pprofProfile()

	loadConfig()
	// 获取
	hashCacheMapper = getCacheHashMap(dirPath)
	resultHashMapper = make(map[string]string)

	err := fs.WalkDir(os.DirFS(dirPath), ".", upsertFile)
	if err != nil {
		panic(err)
	}

	_ = deleteFile()

	_ = writeCache(dirPath)
}

func upsertFile(relatePath string, d fs.DirEntry, err error) error {
	if d.IsDir() {
		return nil
	}
	if needIgnore(relatePath) {
		return nil
	}
	hash, err := utils.FileMD5(fmt.Sprintf("%s/%s", dirPath, relatePath))
	if err != nil {
		panic(err)
	}

	localFilePath := fmt.Sprintf("%s/%s", dirPath, relatePath)
	ossFilePath := fmt.Sprintf("%s/%s", ossPrefix, relatePath)
	cacheHash, exist := hashCacheMapper[relatePath]
	if !exist {
		_, err := pusher.Client.Push(localFilePath, ossFilePath)
		if err != nil {
			panic(err)
		}

		fmt.Printf("推送新文件[%v]到 oss\n", relatePath)
	} else {
		if cacheHash != hash {
			_, err := pusher.Client.Push(localFilePath, ossFilePath)
			if err != nil {
				panic(err)
			}
			fmt.Printf("更新文件[%v]到 oss\n", relatePath)
		}
	}

	resultHashMapper[relatePath] = hash
	delete(hashCacheMapper, relatePath)
	return nil
}

func deleteFile() error {
	for key := range hashCacheMapper {
		// ossFilePath := fmt.Sprintf("%s/%s", ossPrefix, key)
		// pusher.Client.Delete(ossFilePath)
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

func writeCache(dirPath string) error {
	resultBuilder := &strings.Builder{}
	// 按文件名排序写入，防止文件频繁变更
	keys := make([]string, 0, len(resultHashMapper))
	for key := range resultHashMapper {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	for i := range keys {
		resultBuilder.WriteString(fmt.Sprintf("%s[@]%s\n", keys[i], resultHashMapper[keys[i]]))
	}

	cacheFile := fmt.Sprintf("%s/%s", dirPath, share.CACHE_FILE_NAME)
	err := os.WriteFile(cacheFile, []byte(resultBuilder.String()), os.FileMode(0777))
	if err != nil {
		return err
	}
	return nil
}
