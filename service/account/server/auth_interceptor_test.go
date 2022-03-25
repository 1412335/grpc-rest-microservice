package server

import (
	"account/client"
	"context"
	"reflect"
	"testing"

	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"google.golang.org/grpc"
)

func TestNewAuthServerInterceptor(t *testing.T) {
	type args struct {
		userSrv             client.UserClient
		authRequiredMethods map[string]bool
		accessibleRoles     map[string][]string
	}
	tests := []struct {
		name string
		args args
		want *AuthServerInterceptor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAuthServerInterceptor(tt.args.userSrv, tt.args.authRequiredMethods, tt.args.accessibleRoles); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAuthServerInterceptor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthServerInterceptor_Log(t *testing.T) {
	tests := []struct {
		name string
		a    *AuthServerInterceptor
		want log.Factory
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Log(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthServerInterceptor.Log() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthServerInterceptor_Unary(t *testing.T) {
	tests := []struct {
		name string
		a    *AuthServerInterceptor
		want grpc.UnaryServerInterceptor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Unary(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthServerInterceptor.Unary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthServerInterceptor_Stream(t *testing.T) {
	tests := []struct {
		name string
		a    *AuthServerInterceptor
		want grpc.StreamServerInterceptor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Stream(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthServerInterceptor.Stream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthServerInterceptor_authorize(t *testing.T) {
	type args struct {
		ctx    context.Context
		method string
	}
	tests := []struct {
		name    string
		a       *AuthServerInterceptor
		args    args
		want    *client.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.authorize(tt.args.ctx, tt.args.method)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServerInterceptor.authorize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthServerInterceptor.authorize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthServerInterceptor_UnaryInterceptor(t *testing.T) {
	type args struct {
		ctx     context.Context
		req     interface{}
		info    *grpc.UnaryServerInfo
		handler grpc.UnaryHandler
	}
	tests := []struct {
		name     string
		a        *AuthServerInterceptor
		args     args
		wantResp interface{}
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResp, err := tt.a.UnaryInterceptor(tt.args.ctx, tt.args.req, tt.args.info, tt.args.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServerInterceptor.UnaryInterceptor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("AuthServerInterceptor.UnaryInterceptor() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}

func TestAuthServerInterceptor_StreamInterceptor(t *testing.T) {
	type args struct {
		srv     interface{}
		ss      grpc.ServerStream
		info    *grpc.StreamServerInfo
		handler grpc.StreamHandler
	}
	tests := []struct {
		name    string
		a       *AuthServerInterceptor
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.StreamInterceptor(tt.args.srv, tt.args.ss, tt.args.info, tt.args.handler); (err != nil) != tt.wantErr {
				t.Errorf("AuthServerInterceptor.StreamInterceptor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
