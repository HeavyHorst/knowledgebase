package auth

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type TokenGenerator interface {
	GenerateToken(username, password string, admin bool) (string, error)
	GetSecret() []byte
}

type JWTTokenGenerator struct {
	Method jwt.SigningMethod
	Secret []byte
	Exp    time.Duration
}

func (j *JWTTokenGenerator) GenerateToken(username, password string, admin bool) (string, error) {
	token := jwt.New(j.Method)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(j.Exp).Unix()
	claims["iat"] = time.Now().Unix()
	claims["sub"] = username
	claims["isAdmin"] = admin
	tokenString, err := token.SignedString(append([]byte(password), j.Secret...))
	if err != nil {
		return "", errors.Wrapf(err, "couldn't generate the token")
	}
	return tokenString, nil
}

func (j *JWTTokenGenerator) GetSecret() []byte {
	return j.Secret
}
