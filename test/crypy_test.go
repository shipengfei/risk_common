package test

import (
	"context"
	"testing"

	"gitlab.miliantech.com/risk/base/risk_common/crypt"
)

func TestXxx2(t *testing.T) {
	ctt, err := crypt.PrivateEncrypt(context.Background(), crypt.IdCardPrivateKey, "370481200510126712")
	t.Log(ctt, err, len(ctt))

	ctt, _ = crypt.PublicDecrypt(context.Background(), crypt.IdCardPublicKey, ctt)
	t.Log(ctt)
}
