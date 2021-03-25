package v3

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/structs"
	"github.com/gogo/googleapis/google/rpc"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
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
	ErrMissingID         = errors.BadRequest("MISSING_ID", "id", "Missing user id")
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
			Role:        api_v3.Role_USER.String(),
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
		return nil, ErrMissingID
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

// update user by id
func (u *userServiceImpl) Update(ctx context.Context, req *api_v3.UpdateUserRequest) (*api_v3.UpdateUserResponse, error) {
	if len(req.GetUser().GetId()) == 0 {
		return nil, ErrMissingID
	}
	// response
	rsp := &api_v3.UpdateUserResponse{}
	err := u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user User
		// find user by id
		if err := tx.Where(&User{ID: req.GetUser().GetId()}).First(&user).Error; err == gorm.ErrRecordNotFound {
			return ErrNotFound
		} else if err != nil {
			u.logger.For(ctx).Error("Error find user", zap.Error(err))
			return ErrConnectDB
		}
		// check active
		if !user.Active {
			return errors.BadRequest("not active user", "active", "user not active yet")
		}
		u.logger.For(ctx).Info("mask", zap.Strings("path", req.GetUpdateMask().GetPaths()))
		// If there is no update mask do a regular update
		if req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
			user.Fullname = req.GetUser().GetFullname()
			user.Username = req.GetUser().GetUsername()
			// check email valid
			email := strings.ToLower(req.GetUser().GetEmail())
			if !isValidEmail(email) {
				return ErrInvalidEmail
			}
			user.Email = email
			// hash password
			pwdHashed, err := u.genHash(req.GetUser().GetPassword())
			if err != nil {
				u.logger.For(ctx).Error("Hash password failed", zap.Error(err))
				return ErrHashPassword
			}
			user.Password = pwdHashed
		} else {
			st := structs.New(user)
			in := structs.New(req.GetUser())
			for _, path := range req.GetUpdateMask().GetPaths() {
				if path == "id" {
					return status.Error(codes.InvalidArgument, "cannot update id field")
				}
				if path == "email" {
					email := strings.ToLower(req.GetUser().GetEmail())
					if !isValidEmail(email) {
						return ErrInvalidEmail
					}
					user.Email = email
					continue
				}
				if path == "password" {
					// hash password
					pwdHashed, err := u.genHash(req.GetUser().GetPassword())
					if err != nil {
						u.logger.For(ctx).Error("Hash password failed", zap.Error(err))
						return ErrHashPassword
					}
					user.Password = pwdHashed
					continue
				}
				// This doesn't translate properly if a CustomName setting is used,
				// but none of the fields except ID has that set, so NO WORRIES.
				fname := generator.CamelCase(path)
				field, ok := st.FieldOk(fname)
				if !ok {
					st := status.New(codes.InvalidArgument, "invalid field specified")
					des, err := st.WithDetails(&rpc.BadRequest{
						FieldViolations: []*rpc.BadRequest_FieldViolation{{
							Field:       "update_mask",
							Description: fmt.Sprintf("The user message type does not have a field called %q", path),
						}},
					})
					if err != nil {
						return st.Err()
					}
					return des.Err()
				}
				// set update value
				err := field.Set(in.Field(fname).Value())
				if err != nil {
					return err
				}
			}
		}
		// update user in db
		if err := tx.Save(&user).Error; err != nil && strings.Contains(err.Error(), "idx_users_email") {
			return ErrDuplicateEmail
		} else if err != nil {
			return ErrConnectDB
		}
		// response
		rsp.User = user.sanitize()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rsp, err
}

func (u *userServiceImpl) getUsers(ctx context.Context, req *api_v3.ListUsersRequest) ([]*api_v3.User, error) {
	var users []User
	// build sql statement
	psql := u.dal.GetDatabase().WithContext(ctx)
	if req.GetCreatedSince() != nil {
		psql = psql.Where("created_at >= ?", req.GetCreatedSince())
	}
	if req.GetOlderThen() != nil {
		psql = psql.Where("created_at >= CURRENT_TIMESTAMP - INTERVAL (?)", req.GetOlderThen())
	}
	if req.GetId() != nil {
		psql = psql.Where("id = ?", req.GetId())
	}
	if req.GetUsername() != nil {
		psql = psql.Where("username LIKE '%?%'", req.GetUsername().Value)
	}
	if req.GetFullname() != nil {
		psql = psql.Where("fullname LIKE '%?%'", req.GetFullname().Value)
	}
	if req.GetEmail() != nil {
		psql = psql.Where("email LIKE '%?%'", req.GetEmail().Value)
	}
	if req.GetActive() != nil {
		psql = psql.Where("active = ?", req.GetActive().Value)
	}
	if req.GetRole() != api_v3.Role_GUEST {
		psql = psql.Where("role = ?", req.GetRole().String())
	}
	// exec
	if err := psql.Order("created_at desc").Find(&users).Error; err != nil {
		u.logger.For(ctx).Error("Error find users", zap.Error(err))
		return nil, ErrConnectDB
	}
	// check empty from db
	if len(users) == 0 {
		st := status.New(codes.NotFound, "not found users")
		des, err := st.WithDetails(&rpc.PreconditionFailure{
			Violations: []*rpc.PreconditionFailure_Violation{
				{
					Type:        "USER",
					Subject:     "no users",
					Description: "no users have been found",
				},
			},
		})
		if err != nil {
			return nil, des.Err()
		}
		return nil, st.Err()
	}
	// filter
	rsp := make([]*api_v3.User, len(users))
	for i, user := range users {
		// 	switch {
		// 	case req.GetCreatedSince() != nil && user.CreatedAt.Before(*req.GetCreatedSince()):
		// 		continue
		// 	case req.GetOlderThen() != nil && time.Since(user.CreatedAt) >= *req.GetOlderThen():
		// 		continue
		// 	}
		rsp[i] = user.sanitize()
	}
	return rsp, nil
}

func (u *userServiceImpl) List(ctx context.Context, req *api_v3.ListUsersRequest) (*api_v3.ListUsersResponse, error) {
	users, err := u.getUsers(ctx, req)
	if err != nil {
		return nil, err
	}
	// response
	rsp := &api_v3.ListUsersResponse{
		Users: users,
	}
	return rsp, nil
}

func (u *userServiceImpl) ListStream(req *api_v3.ListUsersRequest, srv api_v3.UserService_ListStreamServer) error {
	users, err := u.getUsers(srv.Context(), req)
	if err != nil {
		return err
	}
	for _, user := range users {
		if err := srv.Send(user); err != nil {
			return err
		}
	}
	return nil
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
	if len(req.GetId()) == 0 {
		return nil, ErrMissingID
	}
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
