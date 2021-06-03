package server

import (
	pb "account/api"
	errorSrv "account/error"
	"account/model"
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/1412335/grpc-rest-microservice/pkg/dal/postgres"
	"github.com/1412335/grpc-rest-microservice/pkg/errors"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
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
			Amount:          15000,
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
		req  *pb.UpdateTransactionRequest
		err  error
	}{
		{
			name: "ErrMissingAccountID",
			ctx:  context.TODO(),
			req:  &pb.UpdateTransactionRequest{},
			err:  errorSrv.ErrMissingAccountID,
		},
		{
			name: "ErrMissingTransactionID",
			ctx:  context.TODO(),
			req: &pb.UpdateTransactionRequest{
				Transaction: &pb.Transaction{
					AccountId: "1",
				},
			},
			err: errorSrv.ErrMissingTransactionID,
		},
		{
			name: "ErrTransactionNotFound",
			ctx:  context.TODO(),
			req: &pb.UpdateTransactionRequest{
				Transaction: &pb.Transaction{
					AccountId: "1",
					Id:        "1",
				},
			},
			err: errorSrv.ErrTransactionNotFound,
		},
		{
			name: "ErrUpdateTransactionID",
			ctx:  context.TODO(),
			req: &pb.UpdateTransactionRequest{
				Transaction: &pb.Transaction{
					AccountId: rspTransCreated[0].AccountId,
					Id:        rspTransCreated[0].Id,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"id"},
				},
			},
			err: errorSrv.ErrUpdateTransactionID,
		},
		{
			name: "ErrUpdateTransactionAccountID",
			ctx:  context.TODO(),
			req: &pb.UpdateTransactionRequest{
				Transaction: &pb.Transaction{
					AccountId: rspTransCreated[0].AccountId,
					Id:        rspTransCreated[0].Id,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"account_id"},
				},
			},
			err: errorSrv.ErrUpdateTransactionAccountID,
		},
		{
			name: "ErrUpdateTransactionUserID",
			ctx:  context.TODO(),
			req: &pb.UpdateTransactionRequest{
				Transaction: &pb.Transaction{
					AccountId: rspTransCreated[0].AccountId,
					Id:        rspTransCreated[0].Id,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"user_id"},
				},
			},
			err: errorSrv.ErrUpdateTransactionUserID,
		},
		{
			name: "ErrUpdateTransactionType",
			ctx:  context.TODO(),
			req: &pb.UpdateTransactionRequest{
				Transaction: &pb.Transaction{
					AccountId: rspTransCreated[0].AccountId,
					Id:        rspTransCreated[0].Id,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"transaction_type"},
				},
			},
			err: errorSrv.ErrUpdateTransactionType,
		},
		{
			name: "ErrUpdateUnknowFields",
			ctx:  context.TODO(),
			req: &pb.UpdateTransactionRequest{
				Transaction: &pb.Transaction{
					AccountId: rspTransCreated[0].AccountId,
					Id:        rspTransCreated[0].Id,
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
			name: "ErrInvalidWithdrawTransactionAmount",
			ctx:  context.TODO(),
			req: &pb.UpdateTransactionRequest{
				Transaction: &pb.Transaction{
					AccountId: rspTransCreated[1].AccountId,
					Id:        rspTransCreated[1].Id,
					Amount:    rspTransCreated[1].Amount + 100000,
				},
			},
			err: errorSrv.ErrInvalidTransactionAmount,
		},
		{
			name: "ErrInvalidTransactionAmount",
			ctx:  context.TODO(),
			req: &pb.UpdateTransactionRequest{
				Transaction: &pb.Transaction{
					AccountId: rspTransCreated[0].AccountId,
					Id:        rspTransCreated[0].Id,
					Amount:    rspTransCreated[0].Amount - 100000,
				},
			},
			err: errorSrv.ErrInvalidTransactionAmount,
		},
		{
			name: "UpdateWithdrawTransSuccess",
			ctx:  context.TODO(),
			req: &pb.UpdateTransactionRequest{
				Transaction: &pb.Transaction{
					AccountId: rspTransCreated[1].AccountId,
					Id:        rspTransCreated[1].Id,
					Amount:    5000,
				},
			},
		},
		{
			name: "UpdateDepositTransSuccess",
			ctx:  context.TODO(),
			req: &pb.UpdateTransactionRequest{
				Transaction: &pb.Transaction{
					AccountId: rspTransCreated[0].AccountId,
					Id:        rspTransCreated[0].Id,
					Amount:    10000,
				},
			},
		},
		{
			name: "UpdateWithdrawTransWithMaskSuccess",
			ctx:  context.TODO(),
			req: &pb.UpdateTransactionRequest{
				Transaction: &pb.Transaction{
					AccountId: rspTransCreated[1].AccountId,
					Id:        rspTransCreated[1].Id,
					Amount:    10000,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"amount"},
				},
			},
		},
		{
			name: "UpdateDepositTransWithMaskSuccess",
			ctx:  context.TODO(),
			req: &pb.UpdateTransactionRequest{
				Transaction: &pb.Transaction{
					AccountId: rspTransCreated[0].AccountId,
					Id:        rspTransCreated[0].Id,
					Amount:    15000,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"amount"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeCreated := time.Now()
			got, err := srv.Update(tt.ctx, tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				require.Equal(t, tt.req.Transaction.Amount, got.Transaction.Amount)
				require.LessOrEqual(t, timeCreated.UTC().Second(), got.Transaction.UpdatedAt.AsTime().Second())
				// get trans
				trans := &model.Transaction{ID: tt.req.Transaction.Id}
				err = srv.getTransactionByID(tt.ctx, trans)
				require.NoError(t, err)
				got.Transaction.UpdatedAt = nil
				if !reflect.DeepEqual(trans.Transform2GRPC(), got.Transaction) {
					t.Errorf("transactionServiceImpl.Update() = %v, want %v", trans.Transform2GRPC(), got.Transaction)
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
				for _, rspTrans := range rspTransCreated {
					if rspTrans.Id == tt.req.Transaction.Id {
						if got.Transaction.TransactionType == pb.TransactionType_DEPOSIT {
							require.Equal(t, accountRsp.Account.Balance+got.Transaction.Amount-rspTrans.Amount, acc.Balance)
						} else {
							require.Equal(t, accountRsp.Account.Balance-got.Transaction.Amount+rspTrans.Amount, acc.Balance)
						}
						rspTrans.Amount = trans.Amount
						break
					}
				}
				require.LessOrEqual(t, timeCreated.UTC().Second(), acc.UpdatedAt.Second())
				accountRsp.Account = acc.Transform2GRPC()
			}
		})
	}
}

func Test_transactionServiceImpl_List(t *testing.T) {
	// create service
	srv := newImplTransactionService(t)

	accSrv := newImplService(t)
	// create account
	reqAccounts := []*pb.CreateAccountRequest{
		{
			UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
			Name:    "test-1",
			Bank:    pb.Bank_ACB,
			Balance: 10000,
		},
		{
			UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
			Name:    "test-2",
			Bank:    pb.Bank_VIB,
			Balance: 50000,
		},
	}
	rspAccCreated := make([]*pb.Account, len(reqAccounts))
	for i, acc := range reqAccounts {
		rsp, err := accSrv.Create(context.TODO(), acc)
		require.NoError(t, err)
		require.NotNil(t, rsp)
		rspAccCreated[i] = rsp.Account
	}

	// create accounts
	reqTrans := []*pb.CreateTransactionRequest{
		{
			UserId:          rspAccCreated[0].UserId,
			AccountId:       rspAccCreated[0].Id,
			Amount:          10000,
			TransactionType: pb.TransactionType_DEPOSIT,
		},
		{
			UserId:          rspAccCreated[0].UserId,
			AccountId:       rspAccCreated[1].Id,
			Amount:          5000,
			TransactionType: pb.TransactionType_WITHDRAW,
		},
		{
			UserId:          rspAccCreated[0].UserId,
			AccountId:       rspAccCreated[1].Id,
			Amount:          15000,
			TransactionType: pb.TransactionType_DEPOSIT,
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
		<-time.After(1 * time.Second)
	}
	sort.Slice(rspTransCreated, func(i, j int) bool {
		return rspTransCreated[i].CreatedAt.AsTime().After(rspTransCreated[j].CreatedAt.AsTime())
	})

	tests := []struct {
		name string
		req  *pb.ListTransactionsRequest
		want []*pb.Transaction
		err  error
	}{
		{
			name: "ErrAccountNotFound",
			req: &pb.ListTransactionsRequest{
				Id: wrapperspb.String("1"),
			},
			err: errorSrv.ErrTransactionNotFound,
		},
		{
			name: "ErrAccountNotFoundID",
			req: &pb.ListTransactionsRequest{
				Id:        wrapperspb.String(rspTransCreated[2].Id),
				AccountId: wrapperspb.String(rspTransCreated[1].AccountId),
			},
			err: errorSrv.ErrTransactionNotFound,
		},
		{
			name: "ErrAccountNotFoundUserID",
			req: &pb.ListTransactionsRequest{
				Id:     wrapperspb.String(rspTransCreated[2].Id),
				UserId: wrapperspb.String("1"),
			},
			err: errorSrv.ErrTransactionNotFound,
		},
		{
			name: "SuccessWithID",
			req: &pb.ListTransactionsRequest{
				Id: wrapperspb.String(rspTransCreated[2].Id),
			},
			want: rspTransCreated[2:],
		},
		{
			name: "SuccessWithAccountID",
			req: &pb.ListTransactionsRequest{
				AccountId: wrapperspb.String(rspTransCreated[0].AccountId),
			},
			want: rspTransCreated[:2],
		},
		{
			name: "SuccessWithUserID",
			req: &pb.ListTransactionsRequest{
				UserId: wrapperspb.String(rspTransCreated[0].UserId),
			},
			want: rspTransCreated,
		},
		{
			name: "SuccessWithAmountMin",
			req: &pb.ListTransactionsRequest{
				AmountMin: wrapperspb.Double(10000),
			},
			want: []*pb.Transaction{rspTransCreated[0], rspTransCreated[2]},
		},
		{
			name: "SuccessWithAmountMax",
			req: &pb.ListTransactionsRequest{
				AmountMax: wrapperspb.Double(5000),
			},
			want: rspTransCreated[1:2],
		},
		{
			name: "SuccessWithCreatedSinceLastMinute",
			req: &pb.ListTransactionsRequest{
				CreatedSince: timestamppb.New(time.Now().Add(-1 * time.Minute)),
			},
			want: rspTransCreated,
		},
		{
			name: "SuccessWithCreatedFromLastMinute",
			req: &pb.ListTransactionsRequest{
				OlderThen: durationpb.New(-1 * time.Minute),
			},
			want: rspTransCreated,
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
				require.NotNil(t, rsp.Transactions)
				require.Len(t, rsp.Transactions, len(tt.want))
				for i, want := range tt.want {
					if !reflect.DeepEqual(rsp.Transactions[i], want) {
						t.Errorf("transactionServiceImpl.List() = %v, want %v", rsp.Transactions[i], want)
					}
				}
			}
		})
	}
}

func Test_transactionServiceImpl_ListStream(t *testing.T) {
	// create service
	srv := newImplTransactionService(t)

	accSrv := newImplService(t)
	// create account
	reqAccounts := []*pb.CreateAccountRequest{
		{
			UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
			Name:    "test-1",
			Bank:    pb.Bank_ACB,
			Balance: 10000,
		},
		{
			UserId:  "4a40f3c8-b699-4716-9334-3ec1330c9aee",
			Name:    "test-2",
			Bank:    pb.Bank_VIB,
			Balance: 50000,
		},
	}
	rspAccCreated := make([]*pb.Account, len(reqAccounts))
	for i, acc := range reqAccounts {
		rsp, err := accSrv.Create(context.TODO(), acc)
		require.NoError(t, err)
		require.NotNil(t, rsp)
		rspAccCreated[i] = rsp.Account
	}

	// create accounts
	reqTrans := []*pb.CreateTransactionRequest{
		{
			UserId:          rspAccCreated[0].UserId,
			AccountId:       rspAccCreated[0].Id,
			Amount:          10000,
			TransactionType: pb.TransactionType_DEPOSIT,
		},
		{
			UserId:          rspAccCreated[0].UserId,
			AccountId:       rspAccCreated[1].Id,
			Amount:          5000,
			TransactionType: pb.TransactionType_WITHDRAW,
		},
		{
			UserId:          rspAccCreated[0].UserId,
			AccountId:       rspAccCreated[1].Id,
			Amount:          15000,
			TransactionType: pb.TransactionType_DEPOSIT,
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
		<-time.After(1 * time.Second)
	}
	sort.Slice(rspTransCreated, func(i, j int) bool {
		return rspTransCreated[i].CreatedAt.AsTime().After(rspTransCreated[j].CreatedAt.AsTime())
	})

	tests := []struct {
		name string
		req  *pb.ListTransactionsRequest
		want []*pb.Transaction
		err  error
	}{
		{
			name: "ErrAccountNotFound",
			req: &pb.ListTransactionsRequest{
				Id: wrapperspb.String("1"),
			},
			err: errorSrv.ErrTransactionNotFound,
		},
		{
			name: "ErrAccountNotFoundID",
			req: &pb.ListTransactionsRequest{
				Id:        wrapperspb.String(rspTransCreated[2].Id),
				AccountId: wrapperspb.String(rspTransCreated[1].AccountId),
			},
			err: errorSrv.ErrTransactionNotFound,
		},
		{
			name: "ErrAccountNotFoundUserID",
			req: &pb.ListTransactionsRequest{
				Id:     wrapperspb.String(rspTransCreated[2].Id),
				UserId: wrapperspb.String("1"),
			},
			err: errorSrv.ErrTransactionNotFound,
		},
		{
			name: "SuccessWithID",
			req: &pb.ListTransactionsRequest{
				Id: wrapperspb.String(rspTransCreated[2].Id),
			},
			want: rspTransCreated[2:],
		},
		{
			name: "SuccessWithAccountID",
			req: &pb.ListTransactionsRequest{
				AccountId: wrapperspb.String(rspTransCreated[0].AccountId),
			},
			want: rspTransCreated[:2],
		},
		{
			name: "SuccessWithUserID",
			req: &pb.ListTransactionsRequest{
				UserId: wrapperspb.String(rspTransCreated[0].UserId),
			},
			want: rspTransCreated,
		},
		{
			name: "SuccessWithAmountMin",
			req: &pb.ListTransactionsRequest{
				AmountMin: wrapperspb.Double(10000),
			},
			want: []*pb.Transaction{rspTransCreated[0], rspTransCreated[2]},
		},
		{
			name: "SuccessWithAmountMax",
			req: &pb.ListTransactionsRequest{
				AmountMax: wrapperspb.Double(5000),
			},
			want: rspTransCreated[1:2],
		},
		{
			name: "SuccessWithCreatedSinceLastMinute",
			req: &pb.ListTransactionsRequest{
				CreatedSince: timestamppb.New(time.Now().Add(-1 * time.Minute)),
			},
			want: rspTransCreated,
		},
		{
			name: "SuccessWithCreatedFromLastMinute",
			req: &pb.ListTransactionsRequest{
				OlderThen: durationpb.New(-1 * time.Minute),
			},
			want: rspTransCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockServer := NewMockTransactionService_ListStreamServer(ctrl)
			if tt.err != nil {
				mockServer.EXPECT().Context()
				err := srv.ListStream(tt.req, mockServer)
				require.ErrorIs(t, err, tt.err)
			} else {
				call := make([]*gomock.Call, len(tt.want)+1)
				call[0] = mockServer.EXPECT().Context()
				for i, want := range tt.want {
					call[i+1] = mockServer.EXPECT().Send(want).Return(nil)
				}
				gomock.InOrder(call...)
				err := srv.ListStream(tt.req, mockServer)
				require.NoError(t, err)
			}
		})
	}
}
