package v3

import (
	"context"
	"strings"

	_ "github.com/fatih/structs"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/postgres"
	"github.com/1412335/grpc-rest-microservice/pkg/errors"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
)

var (
	ErrMissingUsername = errors.BadRequest("MISSING_USERNAME", "username", "Missing username")
	ErrMissingFullname = errors.BadRequest("MISSING_FULLNAME", "fullname", "Missing fullname")
	ErrMissingEmail    = errors.BadRequest("MISSING_EMAIL", "email", "Missing email")
	ErrDuplicateEmail  = errors.BadRequest("DUPLICATE_EMAIL", "email", "A user with this email address already exists")
	ErrInvalidEmail    = errors.BadRequest("INVALID_EMAIL", "email", "The email provided is invalid")
	ErrInvalidPassword = errors.BadRequest("INVALID_PASSWORD", "password", "Password must be at least 8 characters long")
	// ErrIncorrectPassword = errors.Unauthorized("INCORRECT_PASSWORD", "Password wrong")
	// ErrMissingId         = errors.BadRequest("MISSING_ID", "Missing id")
	// ErrMissingToken      = errors.BadRequest("MISSING_TOKEN", "Missing token")

	// ErrConnectDB = errors.InternalServerError("CONNECT_DB", "Connecting to database failed")
	// ErrNotFound  = errors.NotFound("NOT_FOUND", "User not found")

	// ErrTokenGenerated = errors.InternalServerError("TOKEN_GEN_FAILED", "Generate token failed")
	// ErrTokenInvalid   = errors.Unauthorized("TOKEN_INVALID", "Token invalid")
	// ErrTokenNotFound  = errors.BadRequest("TOKEN_NOT_FOUND", "Token not found")
	// ErrTokenExpired   = errors.Unauthorized("TOKEN_EXPIRE", "Token expired")

	// ttlToken   = 24 * time.Hour
)

type UsersHandlerOption func(h *userServiceImpl) error

// func WithAudit(audit *audit.Audit) UsersHandlerOption {
// 	return func(h *userServiceImpl) error {
// 		h.audit = audit
// 		return nil
// 	}
// }

// func WithCacheStore(cache *cache.Cache) UsersHandlerOption {
// 	return func(h *userServiceImpl) error {
// 		h.cache = cache
// 		return nil
// 	}
// }

type userServiceImpl struct {
	dal      *postgres.DataAccessLayer
	logger   log.Factory
	tokenSrv *TokenService
}

var _ api_v3.UserServiceServer = (*userServiceImpl)(nil)

func NewUserService(dal *postgres.DataAccessLayer, logger log.Factory, tokenSrv *TokenService) api_v3.UserServiceServer {
	// migrate model
	if err := dal.GetDatabase().AutoMigrate(&User{}); err != nil {
		logger.Bg().Error("migrate db failed", zap.Error(err))
		return nil
	}
	return &userServiceImpl{
		dal:      dal,
		logger:   logger,
		tokenSrv: tokenSrv,
	}
}

func (u *userServiceImpl) genHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (u *userServiceImpl) compareHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (u *userServiceImpl) Create(ctx context.Context, req *api_v3.CreateUserRequest) (*api_v3.CreateUserResponse, error) {
	// validate request
	if len(req.GetUsername()) == 0 {
		return nil, ErrMissingUsername
	}
	if len(req.GetFullname()) == 0 {
		return nil, ErrMissingUsername
	}
	if !isValidEmail(req.Email) {
		return nil, ErrInvalidEmail
	}
	if !isValidPassword(req.Password) {
		return nil, ErrInvalidPassword
	}

	// hash password
	pwdHashed, err := u.genHash(req.Password)
	if err != nil {
		u.logger.For(ctx).Error("Hash password failed", zap.Error(err))
		return nil, errors.InternalServerError("hash password failed")
	}

	rsp := &api_v3.CreateUserResponse{}

	// create
	err = u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user := &User{
			ID:          uuid.New().String(),
			Username:    req.Username,
			Fullname:    req.Fullname,
			Active:      false,
			Email:       strings.ToLower(req.Email),
			Password:    pwdHashed,
			VerifyToken: pwdHashed[:10],
			Role:        api_v3.Role_GUEST.String(),
		}
		if err := tx.Create(user).Error; err != nil && strings.Contains(err.Error(), "idx_users_email") {
			return ErrDuplicateEmail
		} else if err != nil {
			u.logger.For(ctx).Error("Error connecting from db", zap.Error(err))
			return status.Errorf(codes.Internal, "connecting db failed")
		}

		// create token
		token, err := u.tokenSrv.Generate(user)
		if err != nil {
			u.logger.For(ctx).Error("Error generate token", zap.Error(err))
			return status.Errorf(codes.Internal, "generate token failed")
		}

		rsp.User = user.sanitize()
		rsp.Token = token
		return nil
	})
	if err != nil {
		u.logger.For(ctx).Error("Error create user", zap.Error(err))
		return nil, err
	}
	return rsp, nil
}

func (u *userServiceImpl) Delete(ctx context.Context, req *api_v3.DeleteUserRequest) (*api_v3.DeleteUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func (u *userServiceImpl) Update(ctx context.Context, req *api_v3.UpdateUserRequest) (*api_v3.UpdateUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (u *userServiceImpl) List(ctx context.Context, req *api_v3.ListUsersRequest) (*api_v3.ListUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}

func (u *userServiceImpl) ListStream(req *api_v3.ListUsersRequest, srv api_v3.UserService_ListStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method ListStream not implemented")
}
func (u *userServiceImpl) Login(ctx context.Context, req *api_v3.LoginRequest) (*api_v3.LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}

func (u *userServiceImpl) Logout(ctx context.Context, req *api_v3.LogoutRequest) (*api_v3.LogoutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Logout not implemented")
}

func (u *userServiceImpl) Validate(ctx context.Context, req *api_v3.ValidateRequest) (*api_v3.ValidateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Validate not implemented")
}

// func (u *userServiceImpl) AddUser(ctx context.Context, req *api_v3.User) (*types.Empty, error) {
// 	u.mu.Lock()
// 	defer u.mu.Unlock()

// 	if len(u.users) == 0 && req.GetRole() != api_v3.Role_ADMIN {
// 		st := status.New(codes.InvalidArgument, "first user must be admin")
// 		des, err := st.WithDetails(&rpc.BadRequest{
// 			FieldViolations: []*rpc.BadRequest_FieldViolation{
// 				{
// 					Field:       "role",
// 					Description: "The first user must have role of admin",
// 				},
// 			},
// 		})
// 		if err != nil {
// 			return nil, st.Err()
// 		}
// 		return nil, des.Err()
// 	}

// 	for _, u := range u.users {
// 		if u.GetId() == req.GetId() {
// 			return nil, status.Errorf(codes.FailedPrecondition, "user exists")
// 		}
// 	}

// 	if req.GetCreatedAt() == nil {
// 		now := time.Now()
// 		req.CreatedAt = &now
// 	}

// 	u.users = append(u.users, req)

// 	return new(types.Empty), nil
// }
// func (u *userServiceImpl) ListUsers(req *api_v3.ListUsersRequest, srv api_v3.UserService_ListUsersServer) error {
// 	u.mu.RLock()
// 	defer u.mu.RUnlock()
// 	if len(u.users) == 0 {
// 		st := status.New(codes.NotFound, "not found users")
// 		des, err := st.WithDetails(&rpc.PreconditionFailure{
// 			Violations: []*rpc.PreconditionFailure_Violation{
// 				{
// 					Type:        "USER",
// 					Subject:     "no users",
// 					Description: "no users have been found",
// 				},
// 			},
// 		})
// 		if err != nil {
// 			return des.Err()
// 		}
// 		return st.Err()
// 	}
// 	for _, u := range u.users {
// 		switch {
// 		case req.GetCreatedSince() != nil && u.GetCreatedAt().Before(*req.GetCreatedSince()):
// 			continue
// 		case req.GetOlderThen() != nil && time.Since(*u.GetCreatedAt()) >= *req.GetOlderThen():
// 			continue
// 		}
// 		err := srv.Send(u)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
// func (u *userServiceImpl) ListUsersByRole(req *api_v3.UserRole, srv api_v3.UserService_ListUsersByRoleServer) error {
// 	u.mu.RLock()
// 	defer u.mu.RUnlock()
// 	for _, user := range u.users {
// 		if user.GetRole() == req.GetRole() {
// 			if err := srv.Send(user); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }
// func (u *userServiceImpl) UpdateUser(ctx context.Context, req *api_v3.UpdateUserRequest) (*api_v3.User, error) {
// 	u.mu.Lock()
// 	defer u.mu.Unlock()
// 	var user *api_v3.User
// 	for _, u := range u.users {
// 		if u.GetId() == req.GetUser().GetId() {
// 			user = u
// 		}
// 	}
// 	if user == nil {
// 		return nil, status.Errorf(codes.NotFound, "user not found")
// 	}

// 	// st := structs.New(user)
// 	for _, path := range req.GetUpdateMask().GetPaths() {
// 		if path == "id" {
// 			return nil, status.Errorf(codes.InvalidArgument, "cannot update id")
// 		}
// 	}
// 	u.logger.For(ctx).Info("update_mask", zap.Strings("paths", req.GetUpdateMask().GetPaths()))

// 	return user, nil
// }
