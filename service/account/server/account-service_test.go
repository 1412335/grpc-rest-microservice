package server

import (
	pb "account/api"
	errorSrv "account/error"
	"account/model"
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/postgres"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/stretchr/testify/require"
)

func newServiceError(t *testing.T) pb.AccountServiceServer {
	config := configs.ServiceConfig{
		Database: &configs.Database{
			Host: "abc",
			Port: "1000",
		},
	}
	dal, err := postgres.NewDataAccessLayer(context.Background(), config.Database)
	require.Error(t, err)
	require.Nil(t, dal)
	return nil
}

func newService(t *testing.T) pb.AccountServiceServer {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// service configs
	config := configs.ServiceConfig{
		Database: &configs.Database{
			Host:           "postgres",
			Port:           "5432",
			User:           "root",
			Password:       "root",
			Scheme:         "users",
			MaxIdleConns:   10,
			MaxOpenConns:   100,
			ConnectTimeout: 1 * time.Hour,
		},
	}

	// init postgres
	dal, err := postgres.NewDataAccessLayer(ctx, config.Database)
	fmt.Printf("err: %v", err)
	require.NoError(t, err)
	require.NotNil(t, dal)
	require.NotNil(t, dal.GetDatabase())

	// truncate table
	err = dal.GetDatabase().Exec("TRUNCATE TABLE accounts CASCADE").Error
	require.NoError(t, err)

	// migrate db
	err = dal.GetDatabase().AutoMigrate(
		&model.Account{},
	)
	require.NoError(t, err)

	// create server
	return NewAccountService(dal)
}

func TestNewAccountService(t *testing.T) {
	tests := []struct {
		name    string
		caller  func(t *testing.T) pb.AccountServiceServer
		wantErr bool
	}{
		{
			name:    "ConnectDBFailed",
			caller:  newServiceError,
			wantErr: true,
		},
		{
			name:    "Success",
			caller:  newService,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.caller(t)
			if !tt.wantErr {
				require.NotNil(t, srv)
			}
		})
	}
}

func Test_accountServiceImpl_getAccountByID(t *testing.T) {
	type fields struct {
		dal    *postgres.DataAccessLayer
		logger log.Factory
	}
	type args struct {
		ctx     context.Context
		account *model.Account
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &accountServiceImpl{
				dal:    tt.fields.dal,
				logger: tt.fields.logger,
			}
			if err := u.getAccountByID(tt.args.ctx, tt.args.account); (err != nil) != tt.wantErr {
				t.Errorf("accountServiceImpl.getAccountByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_accountServiceImpl_getAccounts(t *testing.T) {
	type fields struct {
		dal    *postgres.DataAccessLayer
		logger log.Factory
	}
	type args struct {
		ctx context.Context
		req *pb.ListAccountsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*pb.Account
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &accountServiceImpl{
				dal:    tt.fields.dal,
				logger: tt.fields.logger,
			}
			got, err := u.getAccounts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceImpl.getAccounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accountServiceImpl.getAccounts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_accountServiceImpl_getAccountsByUserID(t *testing.T) {
	type fields struct {
		dal    *postgres.DataAccessLayer
		logger log.Factory
	}
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*pb.Account
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &accountServiceImpl{
				dal:    tt.fields.dal,
				logger: tt.fields.logger,
			}
			got, err := u.getAccountsByUserID(tt.args.ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceImpl.getAccountsByUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accountServiceImpl.getAccountsByUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_accountServiceImpl_Create(t *testing.T) {
	// create service
	srv := newService(t)

	tests := []struct {
		name string
		ctx  context.Context
		req  *pb.CreateAccountRequest
		err  error
	}{
		{
			name: "ErrMissingUserID",
			ctx:  context.TODO(),
			req:  &pb.CreateAccountRequest{},
			err:  errorSrv.ErrMissingUserID,
		},
		{
			name: "ErrInvalidAccountBalance",
			ctx:  context.TODO(),
			req: &pb.CreateAccountRequest{
				UserId:  "1",
				Balance: -1,
			},
			err: errorSrv.ErrInvalidAccountBalance,
		},
		{
			name: "Success",
			ctx:  context.TODO(),
			req: &pb.CreateAccountRequest{
				UserId:  "1",
				Name:    "test",
				Bank:    pb.Bank_ACB,
				Balance: 10000,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeCreated := time.Now()
			got, err := srv.Create(tt.ctx, tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got.Account)
				require.Equal(t, tt.req.UserId, got.Account.UserId)
				require.Equal(t, tt.req.Name, got.Account.Name)
				require.Equal(t, tt.req.Bank, got.Account.Bank)
				require.Equal(t, tt.req.Balance, got.Account.Balance)
				require.LessOrEqual(t, timeCreated.UTC().Unix(), got.Account.CreatedAt.Seconds)
			}
		})
	}
}

func Test_accountServiceImpl_Delete(t *testing.T) {
	type fields struct {
		dal    *postgres.DataAccessLayer
		logger log.Factory
	}
	type args struct {
		ctx context.Context
		req *pb.DeleteAccountRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.DeleteAccountResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &accountServiceImpl{
				dal:    tt.fields.dal,
				logger: tt.fields.logger,
			}
			got, err := u.Delete(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceImpl.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accountServiceImpl.Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_accountServiceImpl_Update(t *testing.T) {
	type fields struct {
		dal    *postgres.DataAccessLayer
		logger log.Factory
	}
	type args struct {
		ctx context.Context
		req *pb.UpdateAccountRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.UpdateAccountResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &accountServiceImpl{
				dal:    tt.fields.dal,
				logger: tt.fields.logger,
			}
			got, err := u.Update(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceImpl.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accountServiceImpl.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_accountServiceImpl_List(t *testing.T) {
	type fields struct {
		dal    *postgres.DataAccessLayer
		logger log.Factory
	}
	type args struct {
		ctx context.Context
		req *pb.ListAccountsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.ListAccountsResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &accountServiceImpl{
				dal:    tt.fields.dal,
				logger: tt.fields.logger,
			}
			got, err := u.List(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceImpl.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accountServiceImpl.List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_accountServiceImpl_ListStream(t *testing.T) {
	type fields struct {
		dal    *postgres.DataAccessLayer
		logger log.Factory
	}
	type args struct {
		req       *pb.ListAccountsRequest
		streamSrv pb.AccountService_ListStreamServer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &accountServiceImpl{
				dal:    tt.fields.dal,
				logger: tt.fields.logger,
			}
			if err := u.ListStream(tt.args.req, tt.args.streamSrv); (err != nil) != tt.wantErr {
				t.Errorf("accountServiceImpl.ListStream() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
