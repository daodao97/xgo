package xjwt

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func Test_jwt(t *testing.T) {
	privateKeyBytes, err := os.ReadFile("private.pem")
	if err != nil {
		log.Fatalf("无法读取私钥文件: %v", err)
	}

	publicKeyBytes, err := os.ReadFile("public.pem")
	if err != nil {
		log.Fatalf("无法读取公钥文件: %v", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		log.Fatalf("无法解析私钥: %v", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		log.Fatalf("无法解析公钥: %v", err)
	}

	payload := jwt.MapClaims{
		"sub":  "1234567890",
		"name": "John Doe",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	}

	// 生成token
	tokenString, err := GenerateRsaToken(payload, privateKey)
	if err != nil {
		log.Fatalf("生成token失败: %v", err)
	}
	fmt.Printf("生成的token: %v\n", tokenString)

	// 验证token
	claims, err := VerifyRasToken(tokenString, publicKey)
	if err != nil {
		log.Fatalf("验证token失败: %v", err)
	} else {
		fmt.Printf("Subject: %v\n", claims["sub"])
		fmt.Printf("Name: %v\n", claims["name"])
	}
}
