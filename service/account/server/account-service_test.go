package server

import (
	pb "account/api"
	errorSrv "account/error"
	"account/model"
	"context"
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/postgres"
	"github.com/1412335/grpc-rest-microservice/pkg/errors"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randSeq(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func connectDBError(t *testing.T) *postgres.DataAccessLayer {
	dal, err := postgres.NewDataAccessLayer(context.Background(), &configs.Database{
		Host: "abc",
		Port: "1000",
	})
	require.Error(t, err)
	require.Nil(t, dal)
	return dal
}

func connectDB(t *testing.T) *postgres.DataAccessLayer {
	dal, err := postgres.NewDataAccessLayer(context.Background(), &configs.Database{
		Host:           "localhost",
		Port:           "5432",
		User:           "root",
		Password:       "root",
		Scheme:         "users",
		MaxIdleConns:   10,
		MaxOpenConns:   100,
		ConnectTimeout: 1 * time.Hour,
		Debug:          true,
	})
	require.NoError(t, err)
	require.NotNil(t, dal)
	require.NotNil(t, dal.GetDatabase())

	// migrate db
	err = dal.GetDatabase().AutoMigrate(
		&model.Account{},
		&model.Transaction{},
	)
	require.NoError(t, err)

	// truncate table
	err = dal.GetDatabase().Exec("TRUNCATE TABLE accounts, transactions CASCADE").Error
	require.NoError(t, err)

	// migrate db
	err = dal.GetDatabase().AutoMigrate(
		&model.Account{},
	)
	require.NoError(t, err)

	// create server
	return dal
}

func newImplService(t *testing.T) *accountServiceImpl {
	// create service
	return &accountServiceImpl{
		dal:    connectDB(t),
		logger: log.NewFactory(log.WithLevel("DEBUG")),
	}
}

func TestNewAccountService(t *testing.T) {
	tests := []struct {
		name    string
		caller  func(t *testing.T) *postgres.DataAccessLayer
		wantErr bool
	}{
		{
			name:    "ConnectDBFailed",
			caller:  connectDBError,
			wantErr: true,
		},
		{
			name:    "Success",
			caller:  connectDB,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dal := tt.caller(t)
			if !tt.wantErr {
				require.NotNil(t, dal)
				u := NewAccountService(dal)
				require.NotNil(t, u)
			} else {
				require.Nil(t, dal)
			}
		})
	}
}

func Test_accountServiceImpl_getAccountByID(t *testing.T) {
	// create service
	srv := newImplService(t)
	// create account
	account := &pb.CreateAccountRequest{
		UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
		Name:    "test",
		Bank:    pb.Bank_ACB,
		Balance: 10000,
	}
	accountRsp, err := srv.Create(context.TODO(), account)
	require.NoError(t, err)
	require.NotNil(t, accountRsp)

	tests := []struct {
		name    string
		ctx     context.Context
		account *model.Account
		err     error
	}{
		{
			name: "ErrAccountNotFound",
			ctx:  context.TODO(),
			account: &model.Account{
				ID: "1",
			},
			err: errorSrv.ErrAccountNotFound,
		},
		{
			name: "Success",
			ctx:  context.TODO(),
			account: &model.Account{
				ID:     accountRsp.Account.Id,
				UserID: accountRsp.Account.UserId,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := srv.getAccountByID(tt.ctx, tt.account)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.NotNil(t, tt.account)
				require.Empty(t, tt.account.Name)
				require.Empty(t, tt.account.Bank)
				require.Empty(t, tt.account.Balance)
				require.Empty(t, tt.account.CreatedAt)
				require.Empty(t, tt.account.UpdatedAt)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tt.account)
				if !reflect.DeepEqual(tt.account.Transform2GRPC(), accountRsp.Account) {
					t.Errorf("accountServiceImpl.getAccountByID() = %v, want %v", tt.account.Transform2GRPC(), accountRsp.Account)
				}
			}
		})
	}
}

func Test_accountServiceImpl_getAccounts(t *testing.T) {
	// create service
	srv := newImplService(t)

	// create accounts
	reqAccs := []*pb.CreateAccountRequest{
		{
			UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
			Name:    "test-1",
			Bank:    pb.Bank_ACB,
			Balance: 10000,
		},
		{
			UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
			Name:    "test-2",
			Bank:    pb.Bank_VCB,
			Balance: 20000,
		},
	}
	rspAccCreated := make([]*pb.Account, len(reqAccs))
	for i, acc := range reqAccs {
		rsp, err := srv.Create(context.TODO(), acc)
		require.NoError(t, err)
		require.NotNil(t, rsp.Account)
		require.Equal(t, acc.UserId, rsp.Account.UserId)
		require.Equal(t, acc.Name, rsp.Account.Name)
		require.Equal(t, acc.Bank, rsp.Account.Bank)
		require.Equal(t, acc.Balance, rsp.Account.Balance)
		rspAccCreated[i] = rsp.Account
		<-time.After(1 * time.Second)
	}
	sort.Slice(rspAccCreated, func(i, j int) bool {
		return rspAccCreated[i].CreatedAt.AsTime().After(rspAccCreated[j].CreatedAt.AsTime())
	})

	tests := []struct {
		name string
		ctx  context.Context
		req  *pb.ListAccountsRequest
		want []*pb.Account
		err  error
	}{
		{
			name: "ErrAccountNotFound",
			req: &pb.ListAccountsRequest{
				Id: wrapperspb.String("1"),
			},
			err: errorSrv.ErrAccountNotFound,
		},
		{
			name: "SuccessWithID",
			req: &pb.ListAccountsRequest{
				Id: wrapperspb.String(rspAccCreated[0].Id),
			},
			want: rspAccCreated[:1],
		},
		{
			name: "SuccessWithUserID",
			req: &pb.ListAccountsRequest{
				UserId: wrapperspb.String(rspAccCreated[0].UserId),
			},
			want: rspAccCreated,
		},
		{
			name: "SuccessWithBalanceMin",
			req: &pb.ListAccountsRequest{
				BalanceMin: wrapperspb.Double(15000),
			},
			want: rspAccCreated[:1],
		},
		{
			name: "SuccessWithBalanceMax",
			req: &pb.ListAccountsRequest{
				BalanceMax: wrapperspb.Double(5000),
			},
			err: errorSrv.ErrAccountNotFound,
		},
		{
			name: "SuccessWithName",
			req: &pb.ListAccountsRequest{
				Name: wrapperspb.String("test"),
			},
			want: rspAccCreated,
		},
		{
			name: "SuccessWithCreatedSinceLastMinute",
			req: &pb.ListAccountsRequest{
				CreatedSince: timestamppb.New(time.Now().Add(-1 * time.Minute)),
			},
			want: rspAccCreated,
		},
		{
			name: "SuccessWithCreatedFromLastMinute",
			req: &pb.ListAccountsRequest{
				OlderThen: durationpb.New(-1 * time.Minute),
			},
			want: rspAccCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := srv.getAccounts(tt.ctx, tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp)
			} else {
				require.NoError(t, err)
				require.Len(t, rsp, len(tt.want))
				if !reflect.DeepEqual(rsp, tt.want) {
					t.Errorf("accountServiceImpl.getAccounts() = %v, want %v", rsp, tt.want)
				}
			}
		})
	}
}

func Test_accountServiceImpl_getAccountsByUserID(t *testing.T) {
	// create service
	srv := newImplService(t)

	// create accounts
	reqAccs := []*pb.CreateAccountRequest{
		{
			UserId:  "1",
			Name:    "test-1",
			Bank:    pb.Bank_ACB,
			Balance: 10000,
		},
		{
			UserId:  "2",
			Name:    "test-2.1",
			Bank:    pb.Bank_VCB,
			Balance: 20000,
		},
		{
			UserId:  "2",
			Name:    "test-2.2",
			Bank:    pb.Bank_VIB,
			Balance: 5000,
		},
	}
	rspAccCreated := make([]*pb.Account, len(reqAccs))
	for i, acc := range reqAccs {
		rsp, err := srv.Create(context.TODO(), acc)
		require.NoError(t, err)
		require.NotNil(t, rsp.Account)
		require.Equal(t, acc.UserId, rsp.Account.UserId)
		require.Equal(t, acc.Name, rsp.Account.Name)
		require.Equal(t, acc.Bank, rsp.Account.Bank)
		require.Equal(t, acc.Balance, rsp.Account.Balance)
		rspAccCreated[i] = rsp.Account
		<-time.After(1 * time.Second)
	}
	sort.Slice(rspAccCreated, func(i, j int) bool {
		return rspAccCreated[i].CreatedAt.AsTime().After(rspAccCreated[j].CreatedAt.AsTime())
	})

	tests := []struct {
		name   string
		ctx    context.Context
		userID string
		want   []*pb.Account
		err    error
	}{
		{
			name: "ErrMissingUserID",
			ctx:  context.TODO(),
			err:  errorSrv.ErrMissingUserID,
		},
		{
			name:   "ErrAccountNotFound",
			ctx:    context.TODO(),
			userID: "3",
			err:    errorSrv.ErrAccountNotFound,
		},
		{
			name:   "Success",
			ctx:    context.TODO(),
			userID: "1",
			want:   rspAccCreated[2:],
		},
		{
			name:   "SuccessMulti",
			ctx:    context.TODO(),
			userID: "2",
			want:   rspAccCreated[:2],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := srv.getAccountsByUserID(tt.ctx, tt.userID)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp)
			} else {
				require.NoError(t, err)
				require.Len(t, rsp, len(tt.want))
				if !reflect.DeepEqual(rsp, tt.want) {
					t.Errorf("accountServiceImpl.getAccounts() = %v, want %v", rsp, tt.want)
				}
			}
		})
	}
}

func Test_accountServiceImpl_Create(t *testing.T) {
	// create service
	srv := newImplService(t)

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
			name: "ErrValidateNameMaxLen100",
			ctx:  context.TODO(),
			req: &pb.CreateAccountRequest{
				UserId:  "1",
				Balance: 10000,
				Name:    randSeq(101),
				Bank:    pb.Bank_ACB,
			},
			err: errors.BadRequest("validate failed", map[string]string{"Name": "greater than max"}),
		},
		{
			// default: Bank_VCB
			name: "SuccessWithDefaultBank",
			ctx:  context.TODO(),
			req: &pb.CreateAccountRequest{
				UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
				Balance: 100000,
				Name:    randSeq(100),
			},
			// err: errors.BadRequest("validate failed", map[string]string{}),
		},
		{
			name: "Success",
			ctx:  context.TODO(),
			req: &pb.CreateAccountRequest{
				UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
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
				require.LessOrEqual(t, timeCreated.UTC().Second(), got.Account.CreatedAt.AsTime().Second())
			}
		})
	}
}

func Test_accountServiceImpl_Delete(t *testing.T) {
	// create service
	srv := newImplService(t)
	// create account
	account := &pb.CreateAccountRequest{
		UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
		Name:    "test",
		Bank:    pb.Bank_ACB,
		Balance: 10000,
	}
	accountRsp, err := srv.Create(context.TODO(), account)
	require.NoError(t, err)
	require.NotNil(t, accountRsp)

	tests := []struct {
		name string
		ctx  context.Context
		req  *pb.DeleteAccountRequest
		err  error
	}{
		{
			name: "ErrMissingAccountID",
			ctx:  context.TODO(),
			req:  &pb.DeleteAccountRequest{},
			err:  errorSrv.ErrMissingAccountID,
		},
		// {
		// 	name: "ErrAccountNotFound",
		// 	ctx:  context.TODO(),
		// 	req: &pb.DeleteAccountRequest{
		// 		Id: "1",
		// 	},
		// 	err: errorSrv.ErrAccountNotFound,
		// },
		{
			name: "Success",
			ctx:  context.TODO(),
			req: &pb.DeleteAccountRequest{
				Id: accountRsp.Account.Id,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := srv.Delete(tt.ctx, tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				require.Equal(t, tt.req.Id, got.Id)
				err := srv.getAccountByID(tt.ctx, &model.Account{UserID: accountRsp.Account.UserId, ID: accountRsp.Account.Id})
				require.ErrorIs(t, err, errorSrv.ErrAccountNotFound)
			}
		})
	}
}

func Test_accountServiceImpl_Update(t *testing.T) {
	// create service
	srv := newImplService(t)
	// create account
	account := &pb.CreateAccountRequest{
		UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
		Name:    "test",
		Bank:    pb.Bank_ACB,
		Balance: 10000,
	}
	accountRsp, err := srv.Create(context.TODO(), account)
	require.NoError(t, err)
	require.NotNil(t, accountRsp)
	require.NotNil(t, accountRsp.Account)
	require.Equal(t, account.UserId, accountRsp.Account.UserId)
	require.Equal(t, account.Name, accountRsp.Account.Name)
	require.Equal(t, account.Bank, accountRsp.Account.Bank)
	require.Equal(t, account.Balance, accountRsp.Account.Balance)

	tests := []struct {
		name string
		ctx  context.Context
		req  *pb.UpdateAccountRequest
		err  error
	}{
		{
			name: "ErrMissingAccountID",
			ctx:  context.TODO(),
			req: &pb.UpdateAccountRequest{
				Account: &pb.Account{
					Id: "",
				},
			},
			err: errorSrv.ErrMissingAccountID,
		},
		{
			name: "ErrAccountNotFound",
			ctx:  context.TODO(),
			req: &pb.UpdateAccountRequest{
				Account: &pb.Account{
					Id: "1",
				},
			},
			err: errorSrv.ErrAccountNotFound,
		},
		{
			name: "ErrUpdateAccountID",
			ctx:  context.TODO(),
			req: &pb.UpdateAccountRequest{
				Account: &pb.Account{
					Id: accountRsp.Account.Id,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"id"},
				},
			},
			err: errorSrv.ErrUpdateAccountID,
		},
		{
			name: "ErrUpdateAccountUserID",
			ctx:  context.TODO(),
			req: &pb.UpdateAccountRequest{
				Account: &pb.Account{
					Id:     accountRsp.Account.Id,
					UserId: "1",
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"user_id"},
				},
			},
			err: errorSrv.ErrUpdateAccountUserID,
		},
		{
			name: "ErrUpdateAccountBank",
			ctx:  context.TODO(),
			req: &pb.UpdateAccountRequest{
				Account: &pb.Account{
					Id:     accountRsp.Account.Id,
					UserId: "1",
					Bank:   pb.Bank_ACB,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"bank"},
				},
			},
			err: errorSrv.ErrUpdateAccountBank,
		},
		{
			name: "ErrUpdateUnknowFields",
			ctx:  context.TODO(),
			req: &pb.UpdateAccountRequest{
				Account: &pb.Account{
					Id:     accountRsp.Account.Id,
					UserId: "1",
					Bank:   pb.Bank_ACB,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"unknow", "id"},
				},
			},
			err: errors.BadRequest("invalid field specified", map[string]string{
				"update_mask": "account does not have field \"unknow\"",
			}),
		},
		{
			name: "ErrValidate",
			ctx:  context.TODO(),
			req: &pb.UpdateAccountRequest{
				Account: &pb.Account{
					Id:     accountRsp.Account.Id,
					UserId: "1",
					Name:   randSeq(101),
					// Balance: -1,
				},
			},
			err: errors.BadRequest("validate failed", map[string]string{
				"Name": "greater than max",
				// "Balance": "less than min",
			}),
		},
		{
			name: "ErrValidateWithUpdateMask",
			ctx:  context.TODO(),
			req: &pb.UpdateAccountRequest{
				Account: &pb.Account{
					Id:      accountRsp.Account.Id,
					UserId:  "1",
					Name:    randSeq(101),
					Balance: -1,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"balance"},
				},
			},
			err: errors.BadRequest("validate failed", map[string]string{"Balance": "less than min"}),
		},
		{
			name: "ErrValidateWithUpdateMaskMulti",
			ctx:  context.TODO(),
			req: &pb.UpdateAccountRequest{
				Account: &pb.Account{
					Id:      accountRsp.Account.Id,
					UserId:  "1",
					Name:    randSeq(101),
					Balance: -1,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{
						// "balance",
						"name",
					},
				},
			},
			err: errors.BadRequest("validate failed", map[string]string{
				"Name": "greater than max",
				// "Balance": "less than min",
			}),
		},
		{
			name: "Success",
			ctx:  context.TODO(),
			req: &pb.UpdateAccountRequest{
				Account: &pb.Account{
					Id:      accountRsp.Account.Id,
					UserId:  accountRsp.Account.UserId,
					Name:    "test-updated",
					Balance: accountRsp.Account.Balance + 100000,
				},
			},
		},
		{
			name: "SuccessWithUpdateMask",
			ctx:  context.TODO(),
			req: &pb.UpdateAccountRequest{
				Account: &pb.Account{
					Id:      accountRsp.Account.Id,
					UserId:  accountRsp.Account.UserId,
					Name:    "test-updated-2",
					Balance: accountRsp.Account.Balance + 200000,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"balance", "name"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeUpdated := time.Now()
			rsp, err := srv.Update(tt.ctx, tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rsp)
				require.NotNil(t, rsp.Account)
				require.Equal(t, tt.req.Account.UserId, rsp.Account.UserId)
				require.Equal(t, tt.req.Account.Name, rsp.Account.Name)
				require.Equal(t, accountRsp.Account.Bank, rsp.Account.Bank)
				require.Equal(t, tt.req.Account.Balance, rsp.Account.Balance)
				require.LessOrEqual(t, timeUpdated.UTC().Second(), rsp.Account.UpdatedAt.AsTime().Second())
				accountRsp.Account = rsp.Account
			}
		})
	}
}

func Test_accountServiceImpl_List(t *testing.T) {
	// create service
	srv := newImplService(t)

	// create accounts
	reqAccs := []*pb.CreateAccountRequest{
		{
			UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
			Name:    "test-1",
			Bank:    pb.Bank_ACB,
			Balance: 10000,
		},
		{
			UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
			Name:    "test-2",
			Bank:    pb.Bank_VCB,
			Balance: 20000,
		},
	}
	rspAccCreated := make([]*pb.Account, len(reqAccs))
	for i, acc := range reqAccs {
		rsp, err := srv.Create(context.TODO(), acc)
		require.NoError(t, err)
		require.NotNil(t, rsp.Account)
		require.Equal(t, acc.UserId, rsp.Account.UserId)
		require.Equal(t, acc.Name, rsp.Account.Name)
		require.Equal(t, acc.Bank, rsp.Account.Bank)
		require.Equal(t, acc.Balance, rsp.Account.Balance)
		rspAccCreated[i] = rsp.Account
		<-time.After(1 * time.Second)
	}
	sort.Slice(rspAccCreated, func(i, j int) bool {
		return rspAccCreated[i].CreatedAt.AsTime().After(rspAccCreated[j].CreatedAt.AsTime())
	})

	tests := []struct {
		name string
		req  *pb.ListAccountsRequest
		want []*pb.Account
		err  error
	}{
		{
			name: "ErrAccountNotFound",
			req: &pb.ListAccountsRequest{
				Id: wrapperspb.String("1"),
			},
			err: errorSrv.ErrAccountNotFound,
		},
		{
			name: "SuccessWithID",
			req: &pb.ListAccountsRequest{
				Id: wrapperspb.String(rspAccCreated[0].Id),
			},
			want: rspAccCreated[:1],
		},
		{
			name: "SuccessWithUserID",
			req: &pb.ListAccountsRequest{
				UserId: wrapperspb.String(rspAccCreated[0].UserId),
			},
			want: rspAccCreated,
		},
		{
			name: "SuccessWithBalanceMin",
			req: &pb.ListAccountsRequest{
				BalanceMin: wrapperspb.Double(15000),
			},
			want: rspAccCreated[:1],
		},
		{
			name: "SuccessWithBalanceMax",
			req: &pb.ListAccountsRequest{
				BalanceMax: wrapperspb.Double(5000),
			},
			err: errorSrv.ErrAccountNotFound,
		},
		{
			name: "SuccessWithName",
			req: &pb.ListAccountsRequest{
				Name: wrapperspb.String("test"),
			},
			want: rspAccCreated,
		},
		{
			name: "SuccessWithCreatedSinceLastMinute",
			req: &pb.ListAccountsRequest{
				CreatedSince: timestamppb.New(time.Now().Add(-1 * time.Minute)),
			},
			want: rspAccCreated,
		},
		{
			name: "SuccessWithCreatedFromLastMinute",
			req: &pb.ListAccountsRequest{
				OlderThen: durationpb.New(-1 * time.Minute),
			},
			want: rspAccCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := srv.List(context.TODO(), tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rsp.Accounts)
				require.Len(t, rsp.Accounts, len(tt.want))
				if !reflect.DeepEqual(rsp.Accounts, tt.want) {
					t.Errorf("accountServiceImpl.List() = %v, want %v", rsp.Accounts, tt.want)
				}
			}
		})
	}
}
