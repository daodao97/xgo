package xjwt

import (
	"fmt"
	"log"
	"os"
	"testing"

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

	// 生成token
	tokenString, err := generateToken(privateKey)
	if err != nil {
		log.Fatalf("生成token失败: %v", err)
	}
	fmt.Printf("生成的token: %v\n", tokenString)

	// 验证token
	token, err := validateToken(tokenString, publicKey)
	if err != nil {
		log.Fatalf("验证token失败: %v", err)
	}

	if token.Valid {
		fmt.Println("Token 有效")
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			fmt.Printf("Subject: %v\n", claims["sub"])
			fmt.Printf("Name: %v\n", claims["name"])
		}
	} else {
		fmt.Println("Token 无效")
	}
}
