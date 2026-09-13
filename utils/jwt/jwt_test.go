package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	var (
		secret = "123"
		claim  = NewCustomClaims("temp", xid.New().String(), "test", "1s")
	)

	token, err := GenerateToken(claim, secret)
	assert.Nil(t, err)
	assert.NotEmpty(t, token)

	verifiedClaim, err := VerifyToken(token, secret)
	assert.Nil(t, err)
	assert.Equal(t, claim.Category, verifiedClaim.Category)
	assert.Equal(t, claim.Id, verifiedClaim.Id)
	assert.Equal(t, claim.Issuer, verifiedClaim.Issuer)

	time.Sleep(time.Second)

	// 过期 token 必须被拒绝（原断言期望校验成功，属测试设计错误）
	verifiedClaim, err = VerifyToken(token, secret)
	assert.NotNil(t, err)
	assert.Nil(t, verifiedClaim)
}

// 算法混淆回归：VerifyToken 已限定 HS256，异族算法签发的 token 必须拒绝
func TestVerifyTokenRejectsForeignAlg(t *testing.T) {
	secret := "123"
	claim := NewCustomClaims("temp", xid.New().String(), "test", "time.Hour")

	// 用 none 算法构造无签名 token
	unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, claim)
	unsignedToken, err := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.Nil(t, err)

	verifiedClaim, err := VerifyToken(unsignedToken, secret)
	assert.NotNil(t, err)
	assert.Nil(t, verifiedClaim)
}
