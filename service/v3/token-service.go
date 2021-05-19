package v3

import (
	"fmt"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/redis"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/utils"
	"github.com/1412335/grpc-rest-microservice/service/v3/model"

	"go.uber.org/zap"

	"github.com/dgrijalva/jwt-go"
)

type UserClaims struct {
	jwt.StandardClaims
	ID       string `json:"id"`
	Username string `json:"username"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type TokenService struct {
	config     *configs.JWT
	logger     log.Factory
	jwtManager *utils.JWTManager
	redis      *redis.Redis
}

func NewTokenService(config *configs.JWT, redis *redis.Redis) *TokenService {
	return &TokenService{
		config:     config,
		logger:     log.With(zap.String("srv", "token")),
		jwtManager: utils.NewJWTManager(config),
		redis:      redis,
	}
}

func (t *TokenService) Generate(user *model.User) (string, error) {
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
		t.logger.Bg().Error("invalid", zap.Any("claims", claims))
		return nil, fmt.Errorf("invalid token claims")
	}
	return uc, nil
}

func (t *TokenService) getInvalidTokensKey(id string) string {
	return fmt.Sprintf("%s-%s", t.config.InvalidateKey, id)
}

func (t *TokenService) IsInvalidated(id, jwtID string) (bool, error) {
	if t.redis == nil {
		return false, fmt.Errorf("redis store is nil")
	}
	invalidTokens, err := t.redis.LRange(t.getInvalidTokensKey(id))
	if err != nil {
		return false, err
	}
	if invalidTokens == nil {
		return false, fmt.Errorf("missing key")
	}
	for _, token := range invalidTokens.Value.([]string) {
		if token == jwtID {
			return true, nil
		}
	}
	return false, nil
}

func (t *TokenService) Invalidate(id, accessToken string) (bool, error) {
	if t.redis == nil {
		return false, fmt.Errorf("redis store is nil")
	}
	claims, err := t.Verify(accessToken)
	if err != nil {
		return true, err
	}
	if invalid, err := t.IsInvalidated(id, claims.Id); err != nil || invalid {
		return true, err
	}
	if err := t.redis.LPush(&redis.Record{
		Key:    t.getInvalidTokensKey(id),
		Value:  claims.Id,
		Expiry: t.config.InvalidateExpiry,
	}); err != nil {
		return false, err
	}
	return true, nil
}
