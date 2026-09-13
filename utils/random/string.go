package random

import "math/rand/v2"

const (
	fullLetters  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	lowerLetters = "abcdefghijklmnopqrstuvwxyz"
	upperLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func String(n int) string {
	return randString(n, fullLetters)
}

func RandomLowerString(n int) string {
	return randString(n, lowerLetters)
}

func RandomUpperString(n int) string {
	return randString(n, upperLetters)
}

// randString 经 math/rand/v2 全局源（并发安全、自动播种；原实现共享
// *rand.RPC 源并发 race 且 UnixNano 种子可预测）。结果非密码学安全，
// 安全 token 场景请用 crypto/rand。
func randString(n int, letters string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return string(b)
}
