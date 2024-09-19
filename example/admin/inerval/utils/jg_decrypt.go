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
MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAKmiwHqTwyb1uqRn
7pBax3tEfyE3ixb60mtdCPjiDAyDwLuxLeDOEJeYNUsT3SnetUhGxP5P7V2v8FLk
AvCaCe4trOutL6apRu2TFn7sHZclFeiI4xomPMzxTJkKpjQKLaqU0a2hBhG1nJ2M
4y5QldOKbVc2eO2gtqJAivpCD2mHAgMBAAECgYAmieQitPksG72IZlhLkWQqfBhJ
yp2d3eP6IkvMh0ZnfXNG8OzUWtxoJFtPMDcZsRAMWI+emzf5BeSaYFTOpqBEjmjd
kOrjQXbDmQmpjghKQh79x+KRr+smiy5kPqUM8KrocF7Df3VjSL3AUws7Ql15K+PR
rWCMuHovZv9C2JkDyQJBANuMePXfyiPtOiOx/mj1gowioaliJRKuS5Xhpr0r2V0+
MucSqtdgDt3cIpkJh56fZbwBPk1AnyCmOW8Adbt/XosCQQDFzNFAeiNjW5LxKu03
qSrSdajumA5cOIuHukEwCXLw42mpL69Otsv5jfI/UmYR+ltreLU3p5V56QmYdo7p
xhx1AkBNrEr3Ie+P+lPBYS2S0JkZHv92v6RCEavoIOcush66oFC985rBi9h2oXUU
E40Jj3ccpov2JNCnameTX+RHK261AkEAukFziVN5n0XLyGyzk4YoXKWOvZ1RaGWW
fehVGfbL1SlPhZDxcx2OVR/kzNu6YZNuInU3r4COsI1QC9EYIen7QQJAJVE4ICq7
2sxbzTlrhJtuGuBUJ+MYLWV9sjxtyW8LV3rW9O81ikepxjgdmpQXWAPz+1dInwzp
R6shEFi4gcc6vg==
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
		SetBasicAuth("843ed270e97e5959a0a99d99", "80d6195868f9d15b996cc5cd").
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
