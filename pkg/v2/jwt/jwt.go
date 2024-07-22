package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewJwtToken(secret string, data map[string]interface{}) (string, error) {
	return NewJwtTokenExtended(secret, data, 24*60*time.Minute)
}

func NewJwtTokenExtended(secret string, data map[string]interface{}, valid time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	tokenExpirationDuration := valid
	for k, v := range data {
		claims[k] = v
	}
	claims["exp"] = time.Now().Add(tokenExpirationDuration).Unix()
	claims["iat"] = time.Now().Unix()

	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return t, nil
}

func ParseJwtToken(secret string, token string, extract []string) ([]string, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	var result []string
	for _, v := range extract {
		if value, ok := claims[v].(string); ok {
			result = append(result, value)
		} else {
			result = append(result, "")
		}
	}

	return result, nil
}
