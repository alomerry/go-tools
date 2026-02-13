package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

func DecryptAES256CBC(ciphertext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("IV length must equal block size")
	}

	// 创建 CBC 解密器
	mode := cipher.NewCBCDecrypter(block, iv)

	// 解密（密文会被原地修改）
	mode.CryptBlocks(ciphertext, ciphertext)

	// 移除 PKCS5 填充
	return Pkcs5UnPadding(ciphertext), nil
}

// EncryptAES256CBC 使用 AES-256-CBC 加密明文
// 返回: base64编码的IV + 密文, 错误信息
func EncryptAES256CBC(plaintext, key []byte) (string, error) {
	// 1. 创建 AES 密码块
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 2. 对明文进行 PKCS5 填充
	plaintext = Pkcs5Padding([]byte(base64.StdEncoding.EncodeToString(plaintext)), aes.BlockSize)

	// 3. 创建密文字节数组，长度为 IV 长度 + 填充后的明文长度
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	// 4. 生成随机 IV (初始化向量)
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// 5. 创建 CBC 加密器
	mode := cipher.NewCBCEncrypter(block, iv)

	// 6. 执行加密 (将结果放入 ciphertext 的 IV 之后的部分)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	// 7. 返回 Base64 编码的结果
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// EncryptAES256CBCRaw
// 可选的独立加密函数，返回原始字节（IV + 密文）
func EncryptAES256CBCRaw(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext = Pkcs5Padding(plaintext, aes.BlockSize)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}
