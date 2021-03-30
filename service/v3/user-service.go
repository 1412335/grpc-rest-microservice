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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/postgres"
	"github.com/1412335/grpc-rest-microservice/pkg/errors"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/utils"
)

var (
	ErrMissingUsername   = errors.BadRequest("MISSING_USERNAME", map[string]string{"username": "Missing username"})
	ErrMissingFullname   = errors.BadRequest("MISSING_FULLNAME", map[string]string{"fullname": "Missing fullname"})
	ErrMissingEmail      = errors.BadRequest("MISSING_EMAIL", map[string]string{"email": "Missing email"})
	ErrDuplicateEmail    = errors.BadRequest("DUPLICATE_EMAIL", map[string]string{"email": "A user with this email address already exists"})
	ErrInvalidEmail      = errors.BadRequest("INVALID_EMAIL", map[string]string{"email": "The email provided is invalid"})
	ErrInvalidPassword   = errors.BadRequest("INVALID_PASSWORD", map[string]string{"password": "Password must be at least 8 characters long"})
	ErrIncorrectPassword = errors.Unauthenticated("INCORRECT_PASSWORD", "password", "Email or password is incorrect")
	ErrMissingID         = errors.BadRequest("MISSING_ID", map[string]string{"id": "Missing user id"})
	ErrMissingToken      = errors.BadRequest("MISSING_TOKEN", map[string]string{"token": "Missing token"})

	ErrHashPassword = errors.InternalServerError("HASH_PASSWORD", "hash password failed")

	ErrConnectDB = errors.InternalServerError("CONNECT_DB", "Connecting to database failed")
	ErrNotFound  = errors.NotFound("NOT_FOUND", map[string]string{"user": "User not found"})

	ErrTokenGenerated = errors.InternalServerError("TOKEN_GEN_FAILED", "Generate token failed")
	ErrTokenInvalid   = errors.Unauthenticated("TOKEN_INVALID", "token", "Token invalid")
	// ErrTokenNotFound  = errors.BadRequest("TOKEN_NOT_FOUND", "Token not found")
	// ErrTokenExpired   = errors.Unauthorized("TOKEN_EXPIRE", "Token expired")

	ErrUserNotActive = errors.BadRequest("not active user", map[string]string{"active": "user not active yet"})

	// ttlToken   = 24 * time.Hour
)

type userServiceImpl struct {
	dal      *postgres.DataAccessLayer
	logger   log.Factory
	tokenSrv *TokenService
}

var _ api_v3.UserServiceServer = (*userServiceImpl)(nil)

func NewUserService(dal *postgres.DataAccessLayer, tokenSrv *TokenService) api_v3.UserServiceServer {
	return &userServiceImpl{
		dal:      dal,
		logger:   DefaultLogger.With(zap.String("srv", "user")),
		tokenSrv: tokenSrv,
	}
}

// get user by id from redis & db
func (u *userServiceImpl) getUserByID(ctx context.Context, id string) (*User, error) {
	user := &User{}
	err := u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// cache
		if DefaultCache != nil {
			if e := DefaultCache.Get(id, user); e != nil {
				u.logger.For(ctx).Error("Get user cache", zap.Error(e))
			} else {
				return nil
			}
		}
		// find user by id
		if e := tx.Where(&User{ID: id}).First(user).Error; e == gorm.ErrRecordNotFound {
			return ErrNotFound
		} else if e != nil {
			u.logger.For(ctx).Error("Find user", zap.Error(e))
			return ErrConnectDB
		}
		// cache
		if e := user.cache(); e != nil {
			u.logger.For(ctx).Error("Cache user", zap.Error(e))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, err
}

// create user & token
func (u *userServiceImpl) Create(ctx context.Context, req *api_v3.CreateUserRequest) (*api_v3.CreateUserResponse, error) {
	// validate request
	if len(req.GetUsername()) == 0 {
		return nil, ErrMissingUsername
	}
	if len(req.GetFullname()) == 0 {
		return nil, ErrMissingUsername
	}
	if !isValidEmail(req.GetEmail()) {
		return nil, ErrInvalidEmail
	}
	if !isValidPassword(req.GetPassword()) {
		return nil, ErrInvalidPassword
	}

	// init response
	rsp := &api_v3.CreateUserResponse{}

	// create
	return rsp, u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user := &User{
			ID:          uuid.New().String(),
			Username:    req.GetUsername(),
			Fullname:    req.GetFullname(),
			Active:      false,
			Email:       strings.ToLower(req.GetEmail()),
			Password:    req.GetPassword(),
			VerifyToken: "",
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
	rsp := &api_v3.UpdateUserResponse{}
	err := u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// find user by id
		user, e := u.getUserByID(tx.Statement.Context, req.GetUser().GetId())
		if e != nil {
			u.logger.For(ctx).Error("Get user by ID", zap.Error(e))
			return errors.InternalServerError("Get user failed", "Lookup user by ID w redis/db failed")
		}
		// check active
		if !user.Active {
			return ErrUserNotActive
		}
		u.logger.For(ctx).Info("mask", zap.Strings("path", req.GetUpdateMask().GetPaths()))
		// If there is no update mask do a regular update
		if req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
			user.Fullname = req.GetUser().GetFullname()
			user.Username = req.GetUser().GetUsername()
			// check fields valid
			if !isValidEmail(req.GetUser().GetEmail()) {
				return ErrInvalidEmail
			}
			if !isValidPassword(req.GetUser().GetPassword()) {
				return ErrInvalidPassword
			}
		} else {
			st := structs.New(*user)
			in := structs.New(req.GetUser())
			for _, path := range req.GetUpdateMask().GetPaths() {
				if path == "id" {
					return errors.BadRequest("cannot update id", map[string]string{"update_mask": "cannot update id field"})
				}
				// check fields valid
				if path == "email" {
					if !isValidEmail(req.GetUser().GetEmail()) {
						return ErrInvalidEmail
					}
					continue
				}
				if path == "password" {
					if !isValidPassword(req.GetUser().GetPassword()) {
						return ErrInvalidPassword
					}
					continue
				}
				// This doesn't translate properly if a CustomName setting is used,
				// but none of the fields except ID has that set, so NO WORRIES.
				fname := generator.CamelCase(path)
				field, ok := st.FieldOk(fname)
				if !ok {
					return errors.BadRequest("invalid field specified", map[string]string{
						"update_mask": fmt.Sprintf("The user message type does not have a field called %q", path),
					})
				}
				// set update value
				if err := field.Set(in.Field(fname).Value()); err != nil {
					return err
				}
			}
		}
		// update user in db
		if e := tx.Save(user).Error; e != nil && strings.Contains(e.Error(), "idx_users_email") {
			return ErrDuplicateEmail
		} else if e != nil {
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

// build query statement & get list users
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

// list users w unary response
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

// list users w stream response
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

// login w email + pwd & gen token
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
		if e := tx.Where(&User{Email: strings.ToLower(req.GetEmail())}).First(&user).Error; e == gorm.ErrRecordNotFound {
			return ErrNotFound
		} else if e != nil {
			u.logger.For(ctx).Error("Error find user", zap.Error(e))
			return ErrConnectDB
		}
		// verify password
		if e := utils.CompareHash(user.Password, req.GetPassword()); e != nil {
			return ErrIncorrectPassword
		}
		if !user.Active {
			return ErrUserNotActive
		}
		// gen new token
		token, e := u.tokenSrv.Generate(&user)
		if e != nil {
			u.logger.For(ctx).Error("Error gen token", zap.Error(e))
			return ErrTokenGenerated
		}
		// cache user
		if e := user.cache(); e != nil {
			u.logger.For(ctx).Error("Cache user", zap.Error(e))
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

// logout: clear redis cache
func (u *userServiceImpl) Logout(ctx context.Context, req *api_v3.LogoutRequest) (*api_v3.LogoutResponse, error) {
	if len(req.GetId()) == 0 {
		return nil, ErrMissingID
	}
	if DefaultCache != nil {
		if err := DefaultCache.Delete(req.GetId()); err != nil {
			u.logger.For(ctx).Error("logout_clear", zap.Error(err))
			return nil, errors.InternalServerError("logout", "clear cache failed")
		}
	}
	return nil, nil
}

// validate token: update isActive=true & return user
func (u *userServiceImpl) Validate(ctx context.Context, req *api_v3.ValidateRequest) (*api_v3.ValidateResponse, error) {
	if len(req.GetToken()) == 0 {
		return nil, ErrMissingToken
	}
	rsp := &api_v3.ValidateResponse{}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		// verrify token
		claims, e := u.tokenSrv.Verify(req.Token)
		if e != nil {
			u.logger.For(ctx).Error("verify token failed", zap.Error(e))
			return ErrTokenInvalid
		}
		// update active
		if e := tx.Model(&User{ID: claims.ID}).Update("active", true).Error; e == gorm.ErrRecordNotFound {
			return ErrNotFound
		} else if e != nil {
			u.logger.For(ctx).Error("Error update user", zap.Error(e))
			return ErrConnectDB
		}
		// get cache user
		user, e := u.getUserByID(ctx, claims.ID)
		if e != nil {
			u.logger.For(ctx).Error("Get user by ID", zap.Error(e))
			return errors.InternalServerError("Get user failed", "Lookup user by ID w redis/db failed")
		}
		// rsp.User = user.sanitize()
		rsp.Id = claims.ID
		rsp.Username = user.Username
		rsp.Fullname = user.Fullname
		rsp.Email = user.Email
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rsp, err
}
