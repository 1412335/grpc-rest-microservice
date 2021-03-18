package v3

import (
	"context"

	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"
)

type userServiceImpl struct{}

func New() api_v3.UserServiceServer {
	return &userServiceImpl{}
}

func (u *userServiceImpl) AddUser(ctx context.Context, req *api_v3.User) (*types.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddUser not implemented")
}
func (u *userServiceImpl) ListUsers(req *api_v3.ListUsersRequest, srv api_v3.UserService_ListUsersServer) error {
	return status.Errorf(codes.Unimplemented, "method ListUsers not implemented")
}
func (u *userServiceImpl) ListUsersByRole(req *api_v3.UserRole, srv api_v3.UserService_ListUsersByRoleServer) error {
	return status.Errorf(codes.Unimplemented, "method ListUsersByRole not implemented")
}
func (u *userServiceImpl) UpdateUser(ctx context.Context, req *api_v3.UpdateUserRequest) (*api_v3.User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUser not implemented")
}
