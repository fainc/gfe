package test

import (
	"fmt"
	"testing"

	"github.com/fainc/go-crypto/ecdsa"
)

func TestGenECDSAKey(t *testing.T) {
	pri, pub, _ := ecdsa.GenKey()
	fmt.Println(pri.ToBase64String())
	fmt.Println(pub.ToBase64String())
}
