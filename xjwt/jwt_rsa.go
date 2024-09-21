package xjwt

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateRsaToken(payload jwt.MapClaims, privateKey *rsa.PrivateKey) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, payload)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyRasToken(tokenString string, publicKey *rsa.PublicKey) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token.Claims.(jwt.MapClaims), nil
}
