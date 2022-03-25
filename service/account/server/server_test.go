package server

import (
	"reflect"
	"testing"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/server"
)

func TestNewServer(t *testing.T) {
	type args struct {
		srvConfig *configs.ServiceConfig
		opt       []server.Option
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServer(tt.args.srvConfig, tt.args.opt...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_Run(t *testing.T) {
	tests := []struct {
		name    string
		s       *Server
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.Run(); (err != nil) != tt.wantErr {
				t.Errorf("Server.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
