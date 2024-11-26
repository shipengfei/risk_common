package crypt

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
)

var hashPrefixes = map[crypto.Hash][]byte{
	crypto.MD5:       {0x30, 0x20, 0x30, 0x0c, 0x06, 0x08, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x02, 0x05, 0x05, 0x00, 0x04, 0x10},
	crypto.SHA1:      {0x30, 0x21, 0x30, 0x09, 0x06, 0x05, 0x2b, 0x0e, 0x03, 0x02, 0x1a, 0x05, 0x00, 0x04, 0x14},
	crypto.SHA224:    {0x30, 0x2d, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x04, 0x05, 0x00, 0x04, 0x1c},
	crypto.SHA256:    {0x30, 0x31, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x01, 0x05, 0x00, 0x04, 0x20},
	crypto.SHA384:    {0x30, 0x41, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x02, 0x05, 0x00, 0x04, 0x30},
	crypto.SHA512:    {0x30, 0x51, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x03, 0x05, 0x00, 0x04, 0x40},
	crypto.MD5SHA1:   {}, // A special TLS case which doesn't use an ASN1 prefix.
	crypto.RIPEMD160: {0x30, 0x20, 0x30, 0x08, 0x06, 0x06, 0x28, 0xcf, 0x06, 0x03, 0x00, 0x31, 0x04, 0x14},
}

type KeyType int32

const publicKeyType KeyType = 0
const privateKeyType KeyType = 1

var IdCardPublicKey *rsa.PublicKey
var IdCardPrivateKey *rsa.PrivateKey

//go:embed key/rsa_public_key.pem
var idCardPublicKeyBytes []byte

//go:embed key/rsa_private_key.pem
var IdCardPrivateKeyBytes []byte

func init() {
	pubKey, errPub := initKey(publicKeyType, idCardPublicKeyBytes)
	if errPub != nil {
		panic(errPub)
	}
	IdCardPublicKey = pubKey.(*rsa.PublicKey)
	privateKey, errPriv := initKey(privateKeyType, IdCardPrivateKeyBytes)
	if errPriv != nil {
		panic(errPriv)
	}
	IdCardPrivateKey = privateKey.(*rsa.PrivateKey)
}

func initKey(typ KeyType, ctt []byte) (any, error) {
	block, _ := pem.Decode(ctt)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	if typ == publicKeyType {
		return x509.ParsePKIXPublicKey(block.Bytes)
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// 私钥
func PrivateEncrypt(ctx context.Context, key *rsa.PrivateKey, data string) (result string, err error) {
	if data == "" {
		return "", errors.New("encrypt data is null")
	}
	dataByte := []byte(data)
	lenByte := len(dataByte)

	if lenByte <= 117 {
		signData, err := rsa.SignPKCS1v15(nil, key, crypto.Hash(0), []byte(data))
		if err != nil {
			return "", err
		}
		splitAndInsertChar(base64.StdEncoding.EncodeToString(signData), 60, &result)
		return result, nil
	}

	var tempResult string
	runeArray := []rune(data)
	start, end := 0, 0
	for start < len(runeArray) {
		if (start + 29) < len(runeArray) {
			end = start + 29
		} else {
			end = len(runeArray)
		}
		tempRes, err := rsa.SignPKCS1v15(nil, key, crypto.Hash(0), []byte(string(runeArray[start:end])))
		if err != nil {
			return "", err
		}
		item := ""
		splitAndInsertChar(base64.StdEncoding.EncodeToString(tempRes), 60, &item)
		tempResult += item
		if end == len(runeArray) {
			break
		}
		start = end
	}
	return result, nil
}

// 公钥
func PublicDecrypt(ctx context.Context, key *rsa.PublicKey, data string) (string, error) {
	if data == "" {
		return "", errors.New("decrypt data is null")
	}
	// 每175个字符解密一次
	arrayByte := []byte(data)
	lenByte := len(arrayByte)
	start, end := 0, 0
	var resBytes []byte
	for start < lenByte {
		if (start + 175) < lenByte {
			end = start + 175
		} else {
			end = lenByte - 1
		}
		tempByte, err := base64.StdEncoding.DecodeString(string(arrayByte[start:end]))
		if err != nil {
			return "", err
		}
		item, err := publicDecrypt(key, crypto.Hash(0), nil, tempByte)
		if err != nil {
			return "", err
		}
		resBytes = append(resBytes, item...)
		if end == lenByte-1 {
			break
		}
		start = end
	}
	return string(resBytes), nil
}

func publicDecrypt(pub *rsa.PublicKey, hash crypto.Hash, hashed []byte, sig []byte) (out []byte, err error) {
	hashLen, prefix, err := pkcs1v15HashInfo(hash, len(hashed))
	if err != nil {
		return nil, err
	}

	tLen := len(prefix) + hashLen
	k := (pub.N.BitLen() + 7) / 8
	if k < tLen+11 {
		return nil, fmt.Errorf("length illegal")
	}

	c := new(big.Int).SetBytes(sig)
	m := encrypt(new(big.Int), pub, c)
	em := leftPad(m.Bytes(), k)
	out = unLeftPad(em)

	err = nil
	return
}

func encrypt(c *big.Int, pub *rsa.PublicKey, m *big.Int) *big.Int {
	e := big.NewInt(int64(pub.E))
	c.Exp(m, e, pub.N)
	return c
}

// copy from crypt/rsa/pkcs1v5.go
func leftPad(input []byte, size int) (out []byte) {
	n := len(input)
	if n > size {
		n = size
	}
	out = make([]byte, size)
	copy(out[len(out)-n:], input)
	return
}
func unLeftPad(input []byte) (out []byte) {
	n := len(input)
	t := 2
	for i := 2; i < n; i++ {
		if input[i] == 0xff {
			t = t + 1
		} else {
			if input[i] == input[0] {
				t = t + int(input[1])
			}
			break
		}
	}
	out = make([]byte, n-t)
	copy(out, input[t:])
	return
}

func pkcs1v15HashInfo(hash crypto.Hash, inLen int) (hashLen int, prefix []byte, err error) {
	if hash == 0 {
		return inLen, nil, nil
	}

	hashLen = hash.Size()
	if inLen != hashLen {
		return 0, nil, errors.New("crypto/rsa: input must be hashed message")
	}
	prefix, ok := hashPrefixes[hash]
	if !ok {
		return 0, nil, errors.New("crypto/rsa: unsupported hash function")
	}
	return
}

func splitAndInsertChar(key string, num int, temp *string) {
	if len(key) <= num {
		*temp = *temp + key + "\n"
	}
	for i := 0; i < len(key); i++ {
		if (i+1)%num == 0 {
			*temp = *temp + key[:i+1] + "\n"
			key = key[i+1:]
			splitAndInsertChar(key, num, temp)
			break
		}
	}
}
