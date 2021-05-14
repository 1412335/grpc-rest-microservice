package v3

import (
	"fmt"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/redis"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/utils"
	"go.uber.org/zap"

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
	config     *configs.JWT
	logger     log.Factory
	jwtManager *utils.JWTManager
}

func NewTokenService(config *configs.JWT) *TokenService {
	return &TokenService{
		config:     config,
		logger:     log.With(zap.String("srv", "token")),
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
		t.logger.Bg().Error("verify token failed", zap.Error(err))
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
	if DefaultRedisStore == nil {
		return false, fmt.Errorf("RedisStore is nil")
	}
	invalidTokens, err := DefaultRedisStore.LRange(t.getInvalidTokensKey(id))
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
	if DefaultRedisStore == nil {
		return false, fmt.Errorf("RedisStore is nil")
	}
	claims, err := t.Verify(accessToken)
	if err != nil {
		return true, err
	}
	if invalid, err := t.IsInvalidated(id, claims.Id); err != nil || invalid {
		return true, err
	}
	if err := DefaultRedisStore.LPush(&redis.Record{
		Key:    t.getInvalidTokensKey(id),
		Value:  claims.Id,
		Expiry: t.config.InvalidateExpiry,
	}); err != nil {
		return false, err
	}
	return true, nil
}
