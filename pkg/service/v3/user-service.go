package v3

import (
	"context"
	"sync"
	"time"

	_ "github.com/fatih/structs"
	"github.com/gogo/googleapis/google/rpc"
	_ "github.com/gogo/googleapis/google/rpc"
	"github.com/gogo/protobuf/types"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
)

type userServiceImpl struct {
	logger log.Factory
	mu     *sync.RWMutex
	users  []*api_v3.User
}

var _ api_v3.UserServiceServer = (*userServiceImpl)(nil)

func NewUserService(logger log.Factory) api_v3.UserServiceServer {
	return &userServiceImpl{
		logger: logger,
	}
}

func (u *userServiceImpl) AddUser(ctx context.Context, req *api_v3.User) (*types.Empty, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if len(u.users) == 0 && req.GetRole() != api_v3.Role_ADMIN {
		st := status.New(codes.InvalidArgument, "first user must be admin")
		des, err := st.WithDetails(&rpc.BadRequest{
			FieldViolations: []*rpc.BadRequest_FieldViolation{
				{
					Field:       "role",
					Description: "The first user must have role of admin",
				},
			},
		})
		if err != nil {
			return nil, st.Err()
		}
		return nil, des.Err()
	}

	for _, u := range u.users {
		if u.GetId() == req.GetId() {
			return nil, status.Errorf(codes.FailedPrecondition, "user exists")
		}
	}

	if req.GetCreatedAt() == nil {
		now := time.Now()
		req.CreatedAt = &now
	}

	u.users = append(u.users, req)

	return new(types.Empty), nil
}
func (u *userServiceImpl) ListUsers(req *api_v3.ListUsersRequest, srv api_v3.UserService_ListUsersServer) error {
	u.mu.RLock()
	defer u.mu.RUnlock()
	if len(u.users) == 0 {
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
			return des.Err()
		}
		return st.Err()
	}
	for _, u := range u.users {
		switch {
		case req.GetCreatedSince() != nil && u.GetCreatedAt().Before(*req.GetCreatedSince()):
			continue
		case req.GetOlderThen() != nil && time.Since(*u.GetCreatedAt()) >= *req.GetOlderThen():
			continue
		}
		err := srv.Send(u)
		if err != nil {
			return err
		}
	}
	return nil
}
func (u *userServiceImpl) ListUsersByRole(req *api_v3.UserRole, srv api_v3.UserService_ListUsersByRoleServer) error {
	u.mu.RLock()
	defer u.mu.RUnlock()
	for _, user := range u.users {
		if user.GetRole() == req.GetRole() {
			if err := srv.Send(user); err != nil {
				return err
			}
		}
	}
	return nil
}
func (u *userServiceImpl) UpdateUser(ctx context.Context, req *api_v3.UpdateUserRequest) (*api_v3.User, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	var user *api_v3.User
	for _, u := range u.users {
		if u.GetId() == req.GetUser().GetId() {
			user = u
		}
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// st := structs.New(user)
	for _, path := range req.GetUpdateMask().GetPaths() {
		if path == "id" {
			return nil, status.Errorf(codes.InvalidArgument, "cannot update id")
		}
	}
	u.logger.For(ctx).Info("updateMask", zap.Strings("paths", req.GetUpdateMask().GetPaths()))

	return user, nil
}
