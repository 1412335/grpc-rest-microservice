package server

import (
	"context"
	"fmt"
	"time"

	pb "account/api"
	errorSrv "account/error"
	"account/model"

	"github.com/1412335/grpc-rest-microservice/pkg/dal/postgres"
	"github.com/1412335/grpc-rest-microservice/pkg/errors"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type transactionServiceImpl struct {
	dal        *postgres.DataAccessLayer
	logger     log.Factory
	accountSrv *accountServiceImpl
}

var _ pb.TransactionServiceServer = (*transactionServiceImpl)(nil)

func NewTransactionService(dal *postgres.DataAccessLayer, accountSrv *accountServiceImpl) pb.TransactionServiceServer {
	return &transactionServiceImpl{
		dal:        dal,
		accountSrv: accountSrv,
		logger:     log.With(zap.String("srv", "transaction")),
	}
}

// get user by id from redis & db
func (u *transactionServiceImpl) getTransactionByID(ctx context.Context, trans *model.Transaction) error {
	logger := u.logger.With(zap.String("id", trans.ID), zap.String("accountID", trans.AccountID)).For(ctx)
	// get from cache
	if e := trans.GetCache(); e != nil {
		logger.Error("Get trans cache", zap.Error(e))
	} else {
		return nil
	}
	return u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// find in db
		if e := tx.Joins("Account").First(trans).Error; e == gorm.ErrRecordNotFound {
			return errorSrv.ErrTransactionNotFound
		} else if e != nil {
			logger.Error("Lookup trans", zap.Error(e))
			return errorSrv.ErrConnectDB
		}
		// cache
		if e := trans.Cache(); e != nil {
			logger.Error("Cache trans", zap.Error(e))
		}
		return nil
	})
}

// build query statement & get list users
func (u *transactionServiceImpl) getTransactions(ctx context.Context, req *pb.ListTransactionsRequest) ([]*pb.Transaction, error) {
	var transactions []model.Transaction
	// build sql statement
	psql := u.dal.GetDatabase().WithContext(ctx)
	if req.GetAccountId() != nil {
		psql = psql.Where("transactions.account_id = ?", req.GetAccountId().Value)
	}
	if req.GetUserId() != nil {
		psql = psql.Where("\"Account\".\"user_id\" = ?", req.GetUserId().Value)
	}
	if req.GetId() != nil {
		psql = psql.Where("transactions.id = ?", req.GetId().Value)
	}
	if req.GetAmountMin() != nil {
		psql = psql.Where("transactions.amount >= ?", req.GetAmountMin().Value)
	}
	if req.GetAmountMax() != nil {
		psql = psql.Where("transactions.amount <= ?", req.GetAmountMax().Value)
	}
	if req.GetCreatedSince() != nil {
		psql = psql.Where("transactions.created_at >= ?", req.GetCreatedSince().AsTime())
	}
	if req.GetOlderThen() != nil {
		psql = psql.Where("transactions.created_at >= ?", time.Now().Add(req.GetOlderThen().AsDuration()))
	}
	// exec
	if err := psql.Order("transactions.created_at desc").Joins("Account").Find(&transactions).Error; err != nil {
		u.logger.For(ctx).Error("Lookup transactions", zap.Error(err))
		return nil, errorSrv.ErrConnectDB
	}
	// check empty from db
	if len(transactions) == 0 {
		return nil, errorSrv.ErrTransactionNotFound
	}
	// filter
	rsp := make([]*pb.Transaction, len(transactions))
	for i, trans := range transactions {
		rsp[i] = trans.Transform2GRPC()
	}
	return rsp, nil
}

//nolint:unused
func (u *transactionServiceImpl) getTransactionsByAccountID(ctx context.Context, accountID string) ([]*pb.Transaction, error) {
	// // validate request
	// if accountID == "" {
	// 	return nil, errorSrv.ErrMissingAccountID
	// }
	// var transactions []model.Transaction
	// // lookup user by id
	// if e := u.dal.GetDatabase().Where(&model.Transaction{AccountID: accountID}).Order("created_at desc").Find(&transactions).Error; e != nil {
	// 	u.logger.For(ctx).Error("Lookup transactions", zap.String("accountID", accountID), zap.Error(e))
	// 	return nil, errorSrv.ErrConnectDB
	// }
	// if len(transactions) == 0 {
	// 	return nil, errorSrv.ErrTransactionNotFound
	// }
	// // fetch transactions belong to the user
	// rsp := make([]*pb.Transaction, len(transactions))
	// for i, trans := range transactions {
	// 	rsp[i] = trans.Transform2GRPC()
	// }
	// return rsp, nil
	return nil, nil
}

// CreateAccount
func (u *transactionServiceImpl) Create(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
	// validate request
	if req.GetUserId() == "" {
		return nil, errorSrv.ErrMissingUserID
	}
	if req.GetAccountId() == "" {
		return nil, errorSrv.ErrMissingAccountID
	}
	if req.GetAmount() <= 0 {
		return nil, errorSrv.ErrInvalidTransactionAmount
	}

	// response
	rsp := &pb.CreateTransactionResponse{}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		// lookup account
		acc := &model.Account{ID: req.GetAccountId(), UserID: req.GetUserId()}
		if err := u.accountSrv.getAccountByID(tx.Statement.Context, acc); err != nil {
			return err
		}
		// check account balance
		switch req.GetTransactionType() {
		case pb.TransactionType_WITHDRAW:
			acc.Balance -= req.GetAmount()
			if acc.Balance < 0 {
				return errorSrv.ErrInvalidWithdrawTransactionAmount
			}
		case pb.TransactionType_DEPOSIT:
			acc.Balance += req.GetAmount()
		case pb.TransactionType_UNKNOW:
			return errorSrv.ErrUnknowTypeTransaction
		}
		// create transaction
		trans := &model.Transaction{
			ID:              uuid.New().String(),
			AccountID:       req.GetAccountId(),
			TransactionType: req.GetTransactionType().String(),
			Amount:          req.GetAmount(),
			Account:         *acc,
		}
		if err := trans.Validate(); err != nil {
			u.logger.For(ctx).Error("Validate trans", zap.Error(err), zap.Any("details", status.Convert(err).Details()))
			return err
		}
		if err := tx.Omit("Account").Create(trans).Error; err != nil {
			u.logger.For(ctx).Error("Create trans", zap.Any("data", trans), zap.Error(err))
			return errorSrv.ErrConnectDB
		}
		// update account balance
		if err := tx.Select("Balance").Save(&acc).Error; err != nil {
			u.logger.For(ctx).Error("Update account balance", zap.Any("data", acc), zap.Error(err))
			return errorSrv.ErrConnectDB
		}
		//
		rsp.Transaction = trans.Transform2GRPC()
		return nil
	})
	if err != nil {
		return nil, err
	}
	// set header in your handler
	md := metadata.Pairs("X-Http-Code", "201")
	if err := grpc.SetHeader(ctx, md); err != nil {
		u.logger.For(ctx).Error("Set header X-Http-Code", zap.Error(err))
	}
	return rsp, nil
}

func (u *transactionServiceImpl) Delete(ctx context.Context, req *pb.DeleteTransactionRequest) (*pb.DeleteTransactionResponse, error) {
	if req.GetId() == nil {
		return nil, errorSrv.ErrMissingTransactionID
	}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		trans := &model.Transaction{ID: req.GetId().Value}
		if err := u.getTransactionByID(tx.Statement.Context, trans); err != nil {
			return err
		}
		// delete transaction
		if err := tx.Delete(trans).Error; err != nil {
			u.logger.For(ctx).Error("Delete transaction", zap.Error(err))
			return errorSrv.ErrConnectDB
		}
		// increase account balance
		account := trans.Account
		switch trans.TransactionType {
		case pb.TransactionType_WITHDRAW.String():
			account.Balance += trans.Amount
		case pb.TransactionType_DEPOSIT.String():
			account.Balance -= trans.Amount
		case pb.TransactionType_UNKNOW.String():
			return errorSrv.ErrUnknowTypeTransaction
		}
		if err := tx.Select("Balance").Save(&account).Error; err != nil {
			u.logger.For(ctx).Error("Update account balance", zap.Error(err))
			return errorSrv.ErrConnectDB
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &pb.DeleteTransactionResponse{
		Ids: []string{req.GetId().Value},
	}, nil
}

// nolint:gocyclo
func (u *transactionServiceImpl) Update(ctx context.Context, req *pb.UpdateTransactionRequest) (*pb.UpdateTransactionResponse, error) {
	if req.GetTransaction().GetAccountId() == "" {
		return nil, errorSrv.ErrMissingAccountID
	}
	if req.GetTransaction().GetId() == "" {
		return nil, errorSrv.ErrMissingTransactionID
	}

	rsp := &pb.UpdateTransactionResponse{}
	err := u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		trans := &model.Transaction{
			AccountID: req.GetTransaction().GetAccountId(),
			ID:        req.GetTransaction().GetId(),
		}
		// find trans
		if e := u.getTransactionByID(tx.Statement.Context, trans); e != nil {
			return e
		}
		u.logger.For(ctx).Info("mask", zap.Strings("path", req.GetUpdateMask().GetPaths()))
		// If there is no update mask do a regular update
		var paths []string
		if req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
			paths = []string{"amount"}
		} else {
			paths = req.GetUpdateMask().GetPaths()
		}
		for _, path := range paths {
			switch path {
			case "id":
				return errorSrv.ErrUpdateTransactionID
			case "account_id":
				return errorSrv.ErrUpdateTransactionAccountID
			case "user_id":
				return errorSrv.ErrUpdateTransactionUserID
			case "transaction_type":
				return errorSrv.ErrUpdateTransactionType
			case "amount":
				switch trans.TransactionType {
				case pb.TransactionType_WITHDRAW.String():
					trans.Account.Balance += trans.Amount - req.GetTransaction().GetAmount()
				case pb.TransactionType_DEPOSIT.String():
					trans.Account.Balance += -trans.Amount + req.GetTransaction().GetAmount()
				case pb.TransactionType_UNKNOW.String():
					return errorSrv.ErrUnknowTypeTransaction
				}
				if trans.Account.Balance < 0 {
					return errorSrv.ErrInvalidTransactionAmount
				}
				trans.Amount = req.GetTransaction().GetAmount()
			default:
				return errors.BadRequest("invalid field specified", map[string]string{
					"update_mask": fmt.Sprintf("account does not have field %q", path),
				})
			}
		}
		if err := trans.Validate(); err != nil {
			u.logger.For(ctx).Error("Validate trans", zap.Error(err))
			return err
		}
		// update trans
		if err := tx.Select("Amount").Save(trans).Error; err != nil {
			u.logger.For(ctx).Error("Update trans", zap.Error(err))
			return errorSrv.ErrConnectDB
		}
		// update account balance
		if err := tx.Select("Balance").Save(&trans.Account).Error; err != nil {
			u.logger.For(ctx).Error("Update account", zap.Error(err))
			return errorSrv.ErrConnectDB
		}
		// response
		rsp.Transaction = trans.Transform2GRPC()
		rsp.Transaction.UpdatedAt = timestamppb.New(trans.UpdatedAt)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rsp, err
}

func (u *transactionServiceImpl) List(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	transactions, err := u.getTransactions(ctx, req)
	if err != nil {
		return nil, err
	}
	rsp := &pb.ListTransactionsResponse{
		Transactions: transactions,
	}
	return rsp, nil
}

func (u *transactionServiceImpl) ListStream(req *pb.ListTransactionsRequest, streamSrv pb.TransactionService_ListStreamServer) error {
	transactions, err := u.getTransactions(streamSrv.Context(), req)
	if err != nil {
		return err
	}
	for _, trans := range transactions {
		if err := streamSrv.Send(trans); err != nil {
			return err
		}
	}
	return nil
}
