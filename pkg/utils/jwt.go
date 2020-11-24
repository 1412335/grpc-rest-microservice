package utils

import (
	"fmt"
	"time"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"

	"github.com/dgrijalva/jwt-go"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

func NewJWTManager(config *configs.JWT) *JWTManager {
	return &JWTManager{
		secretKey:     config.SecretKey,
		tokenDuration: config.Duration,
	}
}

type UserClaims struct {
	jwt.StandardClaims
	Username string
	Role     string
}

func (manager *JWTManager) Generate(username, role string) (string, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
		},
		Username: username,
		Role:     role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))
}

func (manager *JWTManager) Verify(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return []byte(manager.secretKey), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %+v", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	fmt.Printf("%v %v %v", claims.Username, claims.Role, claims.StandardClaims.ExpiresAt)
	return claims, nil
}
