package v3

import (
	"fmt"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/utils"

	"github.com/dgrijalva/jwt-go"
)

type UserClaims struct {
	jwt.StandardClaims
	ID       string `json:"id"`
	Username string
	Fullname string
	Email    string
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

func (t *TokenService) Generate(user *User) (string, error) {
	claims := UserClaims{
		StandardClaims: t.jwtManager.GetStandardClaims(),
		ID:             user.ID,
		Username:       user.Username,
		Fullname:       user.Fullname,
		Email:          user.Email,
		Role:           user.Role,
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
