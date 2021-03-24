package v2

import (
	"fmt"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/utils"

	"github.com/dgrijalva/jwt-go"
)

type UserClaims struct {
	jwt.StandardClaims
	Username string
	Role     string
}

type TokenService struct {
	jwtManager *utils.JWTManager
}

func NewTokenService(config *configs.JWT) *TokenService {
	return &TokenService{
		jwtManager: utils.NewJWTManager(config),
	}
}

func (t *TokenService) Generate(username, role string) (string, error) {
	claims := UserClaims{
		StandardClaims: t.jwtManager.GetStandardClaims(),
		Username:       username,
		Role:           role,
	}
	return t.jwtManager.Generate(claims)
}

func (t *TokenService) Verify(accessToken string) (*UserClaims, error) {
	claims, err := t.jwtManager.Verify(accessToken, &UserClaims{})
	if err != nil {
		return nil, err
	}
	uc, ok := claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return uc, nil
}
