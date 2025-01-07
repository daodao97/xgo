package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"log"

	"github.com/daodao97/xgo/xdb"
	"github.com/go-resty/resty/v2"
)

var privateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
-----END RSA PRIVATE KEY-----`)

func ParseToken(token string) (string, error) {
	var result = struct {
		Id      int    `json:"id"`
		Code    int    `json:"code"`
		Content string `json:"content"`
		ExID    string `json:"exID"`
		Phone   string `json:"phone"`
	}{}

	resp, err := resty.New().R().
		SetBasicAuth("", "").
		SetBody(xdb.Record{"loginToken": token, "exID": "1234566"}).
		SetResult(&result).
		Post("https://api.verification.jpush.cn/v1/web/loginTokenVerify")

	if err != nil {
		return "", err
	}

	if resp.StatusCode() != 200 {
		return "", errors.New("request failed")
	}

	if result.Code != 8000 {
		return "", errors.New(result.Content)
	}

	return RsaDecrypt(result.Phone)
}

func RsaDecrypt(encrypted string) (string, error) {
	encryptedB, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		log.Println("invalid encrypted")
		return "", err
	}
	result, err := rsaDecrypt(encryptedB, privateKey)
	if err != nil {
		log.Println("err:", err)
		return "", err
	}
	return string(result), nil
}

// 私钥解密
func rsaDecrypt(encrypted, prikey []byte) ([]byte, error) {
	var data []byte
	block, _ := pem.Decode(prikey)
	if block == nil {
		return data, errors.New("private key error")
	}
	rsaKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return data, err
	}
	key, ok := rsaKey.(*rsa.PrivateKey)
	if !ok {
		return data, errors.New("invalid private key")
	}
	data, err = rsa.DecryptPKCS1v15(rand.Reader, key, encrypted)
	return data, err
}
