package crypto

import (
	"bytes"
	"fmt"
)

// Pkcs5UnPadding 移除 PKCS5 填充。带校验：空数据、填充长度为 0 或超出国数据
// 长度均视为非法（原实现无任何校验，伪造填充会静默截坏数据）。
func Pkcs5UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, fmt.Errorf("crypto: empty data")
	}
	unPadding := int(origData[length-1])
	if unPadding == 0 || unPadding > length {
		return nil, fmt.Errorf("crypto: invalid padding size %d", unPadding)
	}
	return origData[:(length - unPadding)], nil
}

func Pkcs5Padding(origData []byte, blockSize int) []byte {
	padding := blockSize - len(origData)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(origData, padText...)
}
