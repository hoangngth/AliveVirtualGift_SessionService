package utils

import (
	"fmt"
	"os"

	"github.com/dgrijalva/jwt-go"
)

//ExtractClaims ...
func ExtractClaims(tokenString string) jwt.MapClaims {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil
		}
		return nil
	}

	if !token.Valid {
		return nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims
	}
	return nil
}
