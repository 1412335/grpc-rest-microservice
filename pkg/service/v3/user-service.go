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
	ErrMissingUsername   = errors.BadRequest("MISSING_USERNAME", "username", "Missing username")
	ErrMissingFullname   = errors.BadRequest("MISSING_FULLNAME", "fullname", "Missing fullname")
	ErrMissingEmail      = errors.BadRequest("MISSING_EMAIL", "email", "Missing email")
	ErrDuplicateEmail    = errors.BadRequest("DUPLICATE_EMAIL", "email", "A user with this email address already exists")
	ErrInvalidEmail      = errors.BadRequest("INVALID_EMAIL", "email", "The email provided is invalid")
	ErrInvalidPassword   = errors.BadRequest("INVALID_PASSWORD", "password", "Password must be at least 8 characters long")
	ErrIncorrectPassword = errors.Unauthenticated("INCORRECT_PASSWORD", "password", "Email or password is incorrect")
	ErrMissingId         = errors.BadRequest("MISSING_ID", "id", "Missing user id")
	ErrMissingToken      = errors.BadRequest("MISSING_TOKEN", "token", "Missing token")

	ErrHashPassword = errors.InternalServerError("HASH_PASSWORD", "hash password failed")

	ErrConnectDB = errors.InternalServerError("CONNECT_DB", "Connecting to database failed")
	ErrNotFound  = errors.NotFound("NOT_FOUND", "user", "User not found")

	ErrTokenGenerated = errors.InternalServerError("TOKEN_GEN_FAILED", "Generate token failed")
	ErrTokenInvalid   = errors.Unauthenticated("TOKEN_INVALID", "token", "Token invalid")
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

// create user
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
		return nil, ErrHashPassword
	}

	// init response
	rsp := &api_v3.CreateUserResponse{}

	// create
	return rsp, u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
			return ErrConnectDB
		}

		// create token
		token, err := u.tokenSrv.Generate(user)
		if err != nil {
			u.logger.For(ctx).Error("Error generate token", zap.Error(err))
			return ErrTokenGenerated
		}

		rsp.User = user.sanitize()
		rsp.Token = token
		return nil
	})
}

// delete user by id
func (u *userServiceImpl) Delete(ctx context.Context, req *api_v3.DeleteUserRequest) (*api_v3.DeleteUserResponse, error) {
	if len(req.GetId()) == 0 {
		return nil, ErrMissingId
	}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(req.GetId()).Delete(&User{}).Error; err == gorm.ErrRecordNotFound {
			return ErrNotFound
		} else if err != nil {
			u.logger.For(ctx).Error("Error connecting from db", zap.Error(err))
			return ErrConnectDB
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &api_v3.DeleteUserResponse{
		Id: req.GetId(),
	}, nil
}

func (u *userServiceImpl) Update(ctx context.Context, req *api_v3.UpdateUserRequest) (*api_v3.UpdateUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (u *userServiceImpl) UpdateV2(ctx context.Context, req *api_v3.UpdateUserRequest) (*api_v3.UpdateUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateV2 not implemented")
}

func (u *userServiceImpl) List(ctx context.Context, req *api_v3.ListUsersRequest) (*api_v3.ListUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}

func (u *userServiceImpl) ListStream(req *api_v3.ListUsersRequest, srv api_v3.UserService_ListStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method ListStream not implemented")
}
func (u *userServiceImpl) Login(ctx context.Context, req *api_v3.LoginRequest) (*api_v3.LoginResponse, error) {
	// validate request
	if len(req.GetEmail()) == 0 {
		return nil, ErrMissingEmail
	}
	if !isValidEmail(req.GetEmail()) {
		return nil, ErrInvalidEmail
	}
	if len(req.GetPassword()) == 0 {
		return nil, ErrInvalidPassword
	}
	// response
	rsp := &api_v3.LoginResponse{}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		var user User
		// find user by email
		if err := tx.Where(&User{Email: strings.ToLower(req.GetEmail())}).First(&user).Error; err == gorm.ErrRecordNotFound {
			return ErrNotFound
		} else if err != nil {
			u.logger.For(ctx).Error("Error find user", zap.Error(err))
			return ErrConnectDB
		}
		// verify password
		if err := u.compareHash(user.Password, req.GetPassword()); err != nil {
			return ErrIncorrectPassword
		}
		if !user.Active {
			return errors.BadRequest("not active user", "active", "user not active yet")
		}
		// gen new token
		token, err := u.tokenSrv.Generate(&user)
		if err != nil {
			u.logger.For(ctx).Error("Error gen token", zap.Error(err))
			return ErrTokenGenerated
		}
		//
		rsp.User = user.sanitize()
		rsp.Token = token
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rsp, err
}

func (u *userServiceImpl) Logout(ctx context.Context, req *api_v3.LogoutRequest) (*api_v3.LogoutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Logout not implemented")
}

func (u *userServiceImpl) Validate(ctx context.Context, req *api_v3.ValidateRequest) (*api_v3.ValidateResponse, error) {
	if len(req.GetToken()) == 0 {
		return nil, ErrMissingToken
	}
	rsp := &api_v3.ValidateResponse{}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		// verrify token
		claims, err := u.tokenSrv.Verify(req.Token)
		if err != nil {
			u.logger.For(ctx).Error("verify token failed", zap.Error(err))
			return ErrTokenInvalid
		}
		// update active
		if err := tx.Model(&User{ID: claims.ID}).Update("active", true).Error; err == gorm.ErrRecordNotFound {
			return ErrNotFound
		} else if err != nil {
			u.logger.For(ctx).Error("Error update user", zap.Error(err))
			return ErrConnectDB
		}
		rsp.Id = claims.ID
		rsp.Username = claims.Username
		rsp.Fullname = claims.Fullname
		rsp.Email = claims.Email
		// rsp.Role = claims.Role
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rsp, err
}

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
