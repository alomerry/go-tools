package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPkcs5PaddingRoundTrip(t *testing.T) {
	// 各种长度（含块对齐边界）填充后可正确还原
	for _, size := range []int{0, 1, 15, 16, 17, 31, 32, 100} {
		data := make([]byte, size)
		_, _ = rand.Read(data)
		padded := Pkcs5Padding(data, 16)
		assert.Equal(t, 0, len(padded)%16)

		unpadded, err := Pkcs5UnPadding(padded)
		assert.Nil(t, err)
		assert.True(t, bytes.Equal(data, unpadded))
	}
}

func TestPkcs5UnPaddingRejectsInvalid(t *testing.T) {
	// 空数据
	_, err := Pkcs5UnPadding(nil)
	assert.NotNil(t, err)

	// 填充字节为 0（非法）
	_, err = Pkcs5UnPadding([]byte{0x00})
	assert.NotNil(t, err)

	// 填充长度超出国数据长度（非法）
	_, err = Pkcs5UnPadding([]byte{0xAA, 0x05})
	assert.NotNil(t, err)
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	// 修复前 Encrypt 先 base64 明文、Decrypt 不解码，二者不互逆；现已互逆
	key := bytes.Repeat([]byte{0x42}, 32) // AES-256 密钥
	plaintext := []byte("hello go-tools 加密 round-trip")

	ciphertext, err := EncryptAES256CBC(plaintext, key)
	assert.Nil(t, err)
	assert.NotEmpty(t, ciphertext)

	// 输出格式为 base64(IV + 密文)，解码后按 IV/密文拆分解密
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	assert.Nil(t, err)
	assert.Greater(t, len(raw), 16)

	decrypted, err := DecryptAES256CBC(raw[16:], key, raw[:16])
	assert.Nil(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestDecryptRejectsBadBlockSize(t *testing.T) {
	// 修复前密文长度非 16 倍数时 CryptBlocks panic；现返回错误
	key := bytes.Repeat([]byte{0x42}, 32)
	iv := bytes.Repeat([]byte{0x01}, 16)

	_, err := DecryptAES256CBC([]byte("not-a-multiple-of-16"), key, iv)
	assert.NotNil(t, err)
}
