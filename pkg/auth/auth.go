package auth

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type TokenGenerator interface {
	GenerateToken(username, password string) (string, error)
	GetSecret() []byte
}

type JWTTokenGenerator struct {
	Method jwt.SigningMethod
	Secret []byte
	Exp    int64
}

func (j *JWTTokenGenerator) GenerateToken(username, password string) (string, error) {
	token := jwt.New(j.Method)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = j.Exp
	claims["iat"] = time.Now().Unix()
	claims["sub"] = username
	tokenString, err := token.SignedString(append([]byte(password), j.Secret...))
	if err != nil {
		return "", errors.Wrapf(err, "couldn't generate the token")
	}
	return tokenString, nil
}

func (j *JWTTokenGenerator) GetSecret() []byte {
	return j.Secret
}
