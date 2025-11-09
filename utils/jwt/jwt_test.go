package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	var (
		secret = "123"
		claim  = NewCustomClaims("temp", uuid.NewString(), "test", "1s")
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

	verifiedClaim, err = VerifyToken(token, secret)
	assert.Nil(t, err)
}
