package xadmin

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/daodao97/xgo/xdb"
)

type JwtConf struct {
	Secret      string
	TokenExpire int64
}

func NewToken(c *JwtConf) *Token {
	return &Token{
		Conf: c,
	}
}

type Token struct {
	Conf     *JwtConf `json:"-"`
	UserID   int      `json:"user_id"`
	Username string   `json:"username"`
	jwt.RegisteredClaims
}

func (t *Token) GenerateToken(id int, email string) (string, error) {
	t.UserID = id
	t.Username = email
	t.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(t.Conf.TokenExpire) * time.Second)), // 过期时间
		IssuedAt:  jwt.NewNumericDate(time.Now()),                                                      // 签发时间
		NotBefore: jwt.NewNumericDate(time.Now()),                                                      // 生效时间
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, t)
	tokenStr, err := token.SignedString([]byte(t.Conf.Secret))

	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (t *Token) ParseToken(tokenStr string) (*Token, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Token{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.Conf.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Token); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

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
