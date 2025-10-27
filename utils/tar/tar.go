package tar

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func UnTar(src, dst string) error {

	// 首先检查文件是否存在和大小
	_, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 判断文件类型并选择相应的解压方法
	if strings.HasSuffix(src, ".tar.gz") || strings.HasSuffix(src, ".tgz") {
		// 解压 tar.gz 文件
		return extractTarGz(src, dst)
	} else if strings.HasSuffix(src, ".gz") {
		// 解压普通 gzip 文件
		return extractGz(src, dst)
	} else if strings.HasSuffix(src, ".tar") {
		// 解压 tar 文件
		return extractTar(src, dst)
	}

	return fmt.Errorf("unsupported file type: %s", src)
}

func extractTarGz(filename, targetDir string) error {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	// 创建 gzip 读取器
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("gzip reader failed: %w", err)
	}
	defer gzReader.Close()

	// 创建 tar 读取器
	tarReader := tar.NewReader(gzReader)

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create directory failed: %w", err)
	}

	// 遍历 tar 文件中的每个文件
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 文件结束
		}
		if err != nil {
			return fmt.Errorf("tar next failed: %w", err)
		}

		// 处理文件路径
		targetPath := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// 创建目录
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("create directory failed: %w", err)
			}

		case tar.TypeReg:
			// 创建文件的父目录
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("create parent directory failed: %w", err)
			}

			file, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("create file failed: %w", err)
			}

			// 复制文件内容
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return fmt.Errorf("copy file content failed: %w", err)
			}
			file.Close()

			// 设置文件权限
			if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("chmod failed: %w", err)
			}

		default:
			fmt.Printf("Skip unsupported type: %c in %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}

// extractGz 解压普通 gzip 文件
func extractGz(filename, targetDir string) error {
	// 打开源文件
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	// 创建 gzip 读取器
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("gzip reader failed: %w", err)
	}
	defer gzReader.Close()

	// 生成目标文件路径（去掉 .gz 后缀）
	baseName := filepath.Base(filename)
	if strings.HasSuffix(baseName, ".gz") {
		baseName = baseName[:len(baseName)-3]
	}

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create directory failed: %w", err)
	}

	// 创建目标文件
	targetPath := filepath.Join(targetDir, baseName)
	outFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("create output file failed: %w", err)
	}
	defer outFile.Close()

	// 复制解压后的内容到目标文件
	if _, err := io.Copy(outFile, gzReader); err != nil {
		return fmt.Errorf("copy file content failed: %w", err)
	}

	return nil
}

// extractTar 解压 tar 文件
func extractTar(filename, targetDir string) error {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	// 创建 tar 读取器
	tarReader := tar.NewReader(file)

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create directory failed: %w", err)
	}

	// 遍历 tar 文件中的每个文件
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 文件结束
		}
		if err != nil {
			return fmt.Errorf("tar next failed: %w", err)
		}

		// 处理文件路径
		targetPath := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// 创建目录
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("create directory failed: %w", err)
			}

		case tar.TypeReg:
			// 创建文件
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("create parent directory failed: %w", err)
			}

			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("create file failed: %w", err)
			}

			// 复制文件内容
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("copy file content failed: %w", err)
			}
			outFile.Close()

			// 设置文件权限
			if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("chmod failed: %w", err)
			}

		default:
			fmt.Printf("Skip unsupported type: %c in %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}
