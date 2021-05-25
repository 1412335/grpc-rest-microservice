package server

import (
	pb "account/api"
	errorSrv "account/error"
	"account/model"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/1412335/grpc-rest-microservice/pkg/dal/postgres"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func newImplTransactionService(t *testing.T) *transactionServiceImpl {
	// create service
	return &transactionServiceImpl{
		dal:        connectDB(t),
		accountSrv: newImplService(t),
		logger:     log.NewFactory(log.WithLevel("DEBUG")),
	}
}

func TestNewTransactionService(t *testing.T) {
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

func Test_transactionServiceImpl_getTransactionByID(t *testing.T) {
	type fields struct {
		dal    *postgres.DataAccessLayer
		logger log.Factory
	}
	type args struct {
		ctx   context.Context
		trans *model.Transaction
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
			u := &transactionServiceImpl{
				dal:    tt.fields.dal,
				logger: tt.fields.logger,
			}
			if err := u.getTransactionByID(tt.args.ctx, tt.args.trans); (err != nil) != tt.wantErr {
				t.Errorf("transactionServiceImpl.getTransactionByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_transactionServiceImpl_getTransactions(t *testing.T) {
	type fields struct {
		dal    *postgres.DataAccessLayer
		logger log.Factory
	}
	type args struct {
		ctx context.Context
		req *pb.ListTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*pb.Transaction
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &transactionServiceImpl{
				dal:    tt.fields.dal,
				logger: tt.fields.logger,
			}
			got, err := u.getTransactions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("transactionServiceImpl.getTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transactionServiceImpl.getTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_transactionServiceImpl_getTransactionsByAccountID(t *testing.T) {
	type fields struct {
		dal    *postgres.DataAccessLayer
		logger log.Factory
	}
	type args struct {
		ctx       context.Context
		accountID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*pb.Transaction
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &transactionServiceImpl{
				dal:    tt.fields.dal,
				logger: tt.fields.logger,
			}
			got, err := u.getTransactionsByAccountID(tt.args.ctx, tt.args.accountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("transactionServiceImpl.getTransactionsByAccountID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transactionServiceImpl.getTransactionsByAccountID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_transactionServiceImpl_Create(t *testing.T) {
	// create service
	srv := newImplTransactionService(t)
	accSrv := newImplService(t)
	// create account
	account := &pb.CreateAccountRequest{
		UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
		Name:    "test",
		Bank:    pb.Bank_ACB,
		Balance: 10000,
	}
	accountRsp, err := accSrv.Create(context.TODO(), account)
	require.NoError(t, err)
	require.NotNil(t, accountRsp)

	tests := []struct {
		name string
		ctx  context.Context
		req  *pb.CreateTransactionRequest
		err  error
	}{
		{
			name: "ErrMissingUserID",
			ctx:  context.TODO(),
			req:  &pb.CreateTransactionRequest{},
			err:  errorSrv.ErrMissingUserID,
		},
		{
			name: "ErrMissingAccountID",
			ctx:  context.TODO(),
			req: &pb.CreateTransactionRequest{
				UserId: "1",
			},
			err: errorSrv.ErrMissingAccountID,
		},
		{
			name: "ErrInvalidTransactionAmount",
			ctx:  context.TODO(),
			req: &pb.CreateTransactionRequest{
				UserId:    "1",
				AccountId: "1",
			},
			err: errorSrv.ErrInvalidTransactionAmount,
		},
		{
			name: "ErrInvalidTransactionAmountNegative",
			ctx:  context.TODO(),
			req: &pb.CreateTransactionRequest{
				UserId:    "1",
				AccountId: "1",
				Amount:    -10000,
			},
			err: errorSrv.ErrInvalidTransactionAmount,
		},
		{
			name: "ErrAccountNotFound",
			ctx:  context.TODO(),
			req: &pb.CreateTransactionRequest{
				UserId:    "1",
				AccountId: "1",
				Amount:    1000,
			},
			err: errorSrv.ErrAccountNotFound,
		},
		{
			name: "ErrInvalidWithdrawTransactionAmount",
			ctx:  context.TODO(),
			req: &pb.CreateTransactionRequest{
				UserId:          accountRsp.Account.UserId,
				AccountId:       accountRsp.Account.Id,
				Amount:          accountRsp.Account.Balance + 1,
				TransactionType: pb.TransactionType_WITHDRAW,
			},
			err: errorSrv.ErrInvalidWithdrawTransactionAmount,
		},
		{
			name: "ErrUnknowTypeTransaction",
			ctx:  context.TODO(),
			req: &pb.CreateTransactionRequest{
				UserId:          accountRsp.Account.UserId,
				AccountId:       accountRsp.Account.Id,
				Amount:          accountRsp.Account.Balance,
				TransactionType: pb.TransactionType_UNKNOW,
			},
			err: errorSrv.ErrUnknowTypeTransaction,
		},
		{
			name: "SuccessDeposit",
			ctx:  context.TODO(),
			req: &pb.CreateTransactionRequest{
				UserId:          accountRsp.Account.UserId,
				AccountId:       accountRsp.Account.Id,
				Amount:          50000,
				TransactionType: pb.TransactionType_DEPOSIT,
			},
		},
		{
			name: "SuccessWithdraw",
			ctx:  context.TODO(),
			req: &pb.CreateTransactionRequest{
				UserId:          accountRsp.Account.UserId,
				AccountId:       accountRsp.Account.Id,
				Amount:          accountRsp.Account.Balance + 15000,
				TransactionType: pb.TransactionType_WITHDRAW,
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
				require.NotNil(t, got.Transaction)
				require.Equal(t, got.Transaction.UserId, tt.req.UserId)
				require.Equal(t, got.Transaction.AccountId, tt.req.AccountId)
				require.Equal(t, got.Transaction.TransactionType, tt.req.TransactionType)
				require.Equal(t, got.Transaction.Amount, tt.req.Amount)
				require.LessOrEqual(t, timeCreated.UTC().Second(), got.Transaction.CreatedAt.AsTime().Second())
				acc := &model.Account{UserID: tt.req.UserId, ID: tt.req.AccountId}
				err = accSrv.getAccountByID(tt.ctx, acc)
				require.NoError(t, err)
				require.NotNil(t, acc)
				require.Equal(t, acc.ID, accountRsp.Account.Id)
				require.Equal(t, acc.UserID, accountRsp.Account.UserId)
				require.Equal(t, acc.Name, accountRsp.Account.Name)
				require.Equal(t, acc.Bank, accountRsp.Account.Bank.String())
				require.True(t, acc.CreatedAt.Equal(accountRsp.Account.CreatedAt.AsTime()))
				if tt.req.TransactionType == pb.TransactionType_DEPOSIT {
					require.Equal(t, acc.Balance, accountRsp.Account.Balance+tt.req.Amount)
				} else {
					require.Equal(t, acc.Balance, accountRsp.Account.Balance-tt.req.Amount)
				}
				require.LessOrEqual(t, timeCreated.UTC().Second(), acc.UpdatedAt.Second())
				accountRsp.Account = acc.Transform2GRPC()
			}
		})
	}
}

func Test_transactionServiceImpl_Delete(t *testing.T) {
	// create service
	srv := newImplTransactionService(t)
	accSrv := newImplService(t)
	// create account
	account := &pb.CreateAccountRequest{
		UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
		Name:    "test",
		Bank:    pb.Bank_ACB,
		Balance: 10000,
	}
	accountRsp, err := accSrv.Create(context.TODO(), account)
	require.NoError(t, err)
	require.NotNil(t, accountRsp)
	// create trans
	reqTrans := []*pb.CreateTransactionRequest{
		{
			UserId:          accountRsp.Account.UserId,
			AccountId:       accountRsp.Account.Id,
			Amount:          5000,
			TransactionType: pb.TransactionType_DEPOSIT,
		},
		{
			UserId:          accountRsp.Account.UserId,
			AccountId:       accountRsp.Account.Id,
			Amount:          5000,
			TransactionType: pb.TransactionType_WITHDRAW,
		},
	}
	rspTransCreated := make([]*pb.Transaction, len(reqTrans))
	for i, trans := range reqTrans {
		rsp, err := srv.Create(context.TODO(), trans)
		require.NoError(t, err)
		require.NotNil(t, rsp.Transaction)
		require.Equal(t, trans.UserId, rsp.Transaction.UserId)
		require.Equal(t, trans.AccountId, rsp.Transaction.AccountId)
		require.Equal(t, trans.Amount, rsp.Transaction.Amount)
		require.Equal(t, trans.TransactionType, rsp.Transaction.TransactionType)
		rspTransCreated[i] = rsp.Transaction
		if rsp.Transaction.TransactionType == pb.TransactionType_DEPOSIT {
			accountRsp.Account.Balance += rsp.Transaction.Amount
		} else {
			accountRsp.Account.Balance -= rsp.Transaction.Amount
		}
	}

	tests := []struct {
		name string
		ctx  context.Context
		req  *pb.Transaction
		err  error
	}{
		{
			name: "ErrMissingTransactionID",
			ctx:  context.TODO(),
			err:  errorSrv.ErrMissingTransactionID,
		},
		{
			name: "ErrTransactionNotFound",
			ctx:  context.TODO(),
			req:  &pb.Transaction{Id: "1"},
			err:  errorSrv.ErrTransactionNotFound,
		},
		{
			name: "SuccessWithdraw",
			ctx:  context.TODO(),
			req:  rspTransCreated[1],
		},
		{
			name: "SuccessDeposit",
			ctx:  context.TODO(),
			req:  rspTransCreated[0],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeCreated := time.Now()
			req := &pb.DeleteTransactionRequest{}
			if tt.req != nil {
				req.Id = wrapperspb.String(tt.req.Id)
			}
			got, err := srv.Delete(tt.ctx, req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				require.Len(t, got.Ids, 1)
				require.Equal(t, got.Ids[0], tt.req.Id)
				// get trans
				trans := &model.Transaction{ID: tt.req.Id}
				err = srv.getTransactionByID(tt.ctx, trans)
				require.ErrorIs(t, err, errorSrv.ErrTransactionNotFound)
				require.Empty(t, trans.AccountID)
				require.Empty(t, trans.Account.ID)
				if !reflect.DeepEqual(trans, &model.Transaction{ID: tt.req.Id}) {
					t.Errorf("accountServiceImpl.List() = %v, want %v", trans, &model.Transaction{ID: tt.req.Id})
				}
				// account
				acc := &model.Account{ID: accountRsp.Account.Id, UserID: accountRsp.Account.UserId}
				err = accSrv.getAccountByID(tt.ctx, acc)
				require.NoError(t, err)
				require.NotNil(t, acc)
				require.Equal(t, acc.ID, accountRsp.Account.Id)
				require.Equal(t, acc.UserID, accountRsp.Account.UserId)
				require.Equal(t, acc.Name, accountRsp.Account.Name)
				require.Equal(t, acc.Bank, accountRsp.Account.Bank.String())
				require.True(t, acc.CreatedAt.Equal(accountRsp.Account.CreatedAt.AsTime()))
				if tt.req.TransactionType == pb.TransactionType_DEPOSIT {
					require.Equal(t, acc.Balance, accountRsp.Account.Balance-tt.req.Amount)
				} else {
					require.Equal(t, acc.Balance, accountRsp.Account.Balance+tt.req.Amount)
				}
				require.LessOrEqual(t, timeCreated.UTC().Second(), acc.UpdatedAt.Second())
				accountRsp.Account = acc.Transform2GRPC()
			}
		})
	}
}

func Test_transactionServiceImpl_Update(t *testing.T) {
	type fields struct {
		dal    *postgres.DataAccessLayer
		logger log.Factory
	}
	type args struct {
		ctx context.Context
		req *pb.UpdateTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.UpdateTransactionResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &transactionServiceImpl{
				dal:    tt.fields.dal,
				logger: tt.fields.logger,
			}
			got, err := u.Update(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("transactionServiceImpl.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transactionServiceImpl.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_transactionServiceImpl_List(t *testing.T) {
	type fields struct {
		dal    *postgres.DataAccessLayer
		logger log.Factory
	}
	type args struct {
		ctx context.Context
		req *pb.ListTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.ListTransactionsResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &transactionServiceImpl{
				dal:    tt.fields.dal,
				logger: tt.fields.logger,
			}
			got, err := u.List(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("transactionServiceImpl.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transactionServiceImpl.List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_transactionServiceImpl_ListStream(t *testing.T) {
	type fields struct {
		dal    *postgres.DataAccessLayer
		logger log.Factory
	}
	type args struct {
		req       *pb.ListTransactionsRequest
		streamSrv pb.TransactionService_ListStreamServer
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
			u := &transactionServiceImpl{
				dal:    tt.fields.dal,
				logger: tt.fields.logger,
			}
			if err := u.ListStream(tt.args.req, tt.args.streamSrv); (err != nil) != tt.wantErr {
				t.Errorf("transactionServiceImpl.ListStream() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
