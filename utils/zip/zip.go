package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Zipper struct {
	dst string
	fw  *os.File
	zw  *zip.Writer
}

func New(dst string) *Zipper {
	return &Zipper{
		dst: dst,
	}
}

func (z *Zipper) FromFiles(srcFiles []string) error {
	var err error
	z.fw, err = os.Create(z.dst)
	if err != nil {
		return err
	}
	z.zw = zip.NewWriter(z.fw)
	for _, filePath := range srcFiles {
		dir, file := filepath.Split(filePath)
		fs, err := os.DirFS(dir).Open(file)
		if err != nil {
			panic(err)
		}
		info, err := fs.Stat()
		if err != nil {
			panic(err)
		}
		fh, err := zip.FileInfoHeader(info)
		if err != nil {
			panic(err)
		}

		// 写入文件信息，并返回一个 Write 结构
		w, err := z.zw.CreateHeader(fh)
		if err != nil {
			panic(err)
		}

		fr, err := os.Open(filePath)
		defer fr.Close()
		if err != nil {
			panic(err)
		}

		// 将打开的文件 Copy 到 w
		n, err := io.Copy(w, fr)
		if err != nil {
			panic(err)
		}
		// 输出压缩的内容
		fmt.Printf("成功压缩文件： %s, 共写入了 %d 个字符的数据\n", filePath, n)
	}
	return nil
}

func (z *Zipper) Close() error {
	err := z.zw.Close()
	if err != nil {
		panic(err)
	}
	return z.fw.Close()
}

//func Zip(dst, src string) (err error) {
//	fw, err := os.Create(dst)
//	defer fw.Close()
//	if err != nil {
//		return err
//	}
//
//	// 通过 fw 来创建 zip.Write
//	zw := zip.NewWriter(fw)
//	defer func() {
//		// 检测一下是否成功关闭
//		if err := zw.Close(); err != nil {
//			log.Fatalln(err)
//		}
//	}()
//
//	// 下面来将文件写入 zw ，因为有可能会有很多个目录及文件，所以递归处理
//	return filepath.Walk(src, func(path string, fi os.FileInfo, errBack error) (err error) {
//		if errBack != nil {
//			return errBack
//		}
//
//		// 通过文件信息，创建 zip 的文件信息
//		fh, err := zip.FileInfoHeader(fi)
//		if err != nil {
//			return
//		}
//
//		// 替换文件信息中的文件名
//		fh.Name = strings.TrimPrefix(path, string(filepath.Separator))
//
//		// 这步开始没有加，会发现解压的时候说它不是个目录
//		if fi.IsDir() {
//			fh.Name += "/"
//		}
//
//		// 写入文件信息，并返回一个 Write 结构
//		w, err := zw.CreateHeader(fh)
//		if err != nil {
//			return
//		}
//
//		// 检测，如果不是标准文件就只写入头信息，不写入文件数据到 w
//		// 如目录，也没有数据需要写
//		if !fh.Mode().IsRegular() {
//			return nil
//		}
//
//		// 打开要压缩的文件
//		fr, err := os.Open(path)
//		defer fr.Close()
//		if err != nil {
//			return
//		}
//
//		// 将打开的文件 Copy 到 w
//		n, err := io.Copy(w, fr)
//		if err != nil {
//			return
//		}
//		// 输出压缩的内容
//		fmt.Printf("成功压缩文件： %s, 共写入了 %d 个字符的数据\n", path, n)
//
//		return nil
//	})
//}

//func Zip(zipName string, files []string) (string, error) {
//	buf := new(bytes.Buffer)
//	zipWriter := zip.NewWriter(buf)
//	defer zipWriter.Close()
//
//	for _, file := range files {
//		f, err := zipWriter.Create(GetFileName(file))
//		if err != nil {
//			log.Fatal(err)
//		}
//		data, err := os.ReadFile(file)
//		if err != nil {
//			return "", err
//		}
//		_, err = zipWriter.Write(data)
//		if err != nil {
//			return "", err
//		}
//	}
//
//	zipWriter.Flush()
//	return zipName, nil
//}

//func GetFile(url string) (resp *http.Response, err error) {
//	u, _ := net_url.Parse(url)
//	if conf.GetString("env") == "picc-production" && u.Host == WEWORK_HOST {
//		url = fmt.Sprintf("%s%s", PICC_WEWORK_HOST, u.Path)
//	}
//	for i := 0; i <= RETRY_TIMES; i++ {
//		resp, err = http.Get(url)
//		if err != nil && strings.Contains(err.Error(), "connection reset by peer") {
//			time.Sleep(500 * time.Millisecond)
//			continue
//		}
//		return
//	}
//	return
//}
