package xjwt

import (
	"github.com/daodao97/xgo/xdb"
	"github.com/golang-jwt/jwt/v5"
)

func JwtPayload(tokenString string) (xdb.Record, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	payload := xdb.Record{}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		for key, value := range claims {
			payload[key] = value
		}
		return payload, nil
	}

	return nil, err
}
