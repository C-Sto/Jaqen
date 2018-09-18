package test

import (
	"crypto/aes"
	"fmt"
	"testing"

	"github.com/c-sto/Jaqen/libJaqen/server/util"
)

func TestEncrypt(t *testing.T) {
	plaintext := "this some plaintext"
	key, e := util.GenerateRandomBytes(32)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}
	ct, e := util.EncryptString(plaintext, key)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}

	testPad, e := util.Pkcs7Pad([]byte("test"), aes.BlockSize)
	if e != nil || testPad[len(testPad)-1] != 12 {
		t.Fail()
	}

	testDecrypt, e := util.Decrypt(ct, key)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}
	if string(testDecrypt) != plaintext {
		fmt.Println(string(testDecrypt))
		t.Fail()
	}
}

//todo: Ensure mac works as expected (modified CT)
//todo: Test bad key
