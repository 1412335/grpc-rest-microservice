package model

import (
	"encoding/json"
	"time"

	pb "account/api"

	"github.com/1412335/grpc-rest-microservice/pkg/cache"
	"github.com/1412335/grpc-rest-microservice/pkg/errors"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

type Transaction struct {
	ID              string    `json:"id"`
	AccountID       string    `json:"account_id" validate:"nonzero"`
	Amount          float64   `json:"amount" validate:"min=1"`
	TransactionType string    `json:"transaction_type"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Account         Account   `json:"account" gorm:"foreignKey:AccountID" `
}

func (a *Transaction) Transform2GRPC() *pb.Transaction {
	data := &pb.Transaction{
		Id:        a.ID,
		AccountId: a.AccountID,
		UserId:    a.Account.UserID,
		Amount:    a.Amount,
		CreatedAt: timestamppb.New(a.CreatedAt),
	}
	if t, ok := pb.TransactionType_value[a.TransactionType]; ok {
		data.TransactionType = pb.TransactionType(t)
	}
	if t, ok := pb.Bank_value[a.Account.Bank]; ok {
		data.Bank = pb.Bank(t)
	}
	return data
}

func (a *Transaction) UpdateFromGRPC(data *pb.Transaction) {
	a.Amount = data.GetAmount()
}

func (a *Transaction) genCacheKey() string {
	return a.AccountID + "_" + a.ID
}

func (a *Transaction) GetCache() error {
	var bytes []byte
	if err := cache.Get(a.genCacheKey(), &bytes); err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, a); err != nil {
		return err
	}
	return nil
}

func (a *Transaction) Cache() error {
	return nil
}

func (a *Transaction) DelCache() error {
	return nil
}

func (a *Transaction) sanitize() {
}

func (a *Transaction) Validate() error {
	// sanitize fileds
	a.sanitize()
	// validate
	e := validator.Validate(a)
	if e != nil {
		errs, ok := e.(validator.ErrorMap)
		if !ok {
			return errors.BadRequest("validate failed", map[string]string{"error": errs.Error()})
		}
		fields := make(map[string]string, len(errs))
		for field, err := range errs {
			fields[field] = err[0].Error()
		}
		return errors.BadRequest("validate failed", fields)
	}
	return nil
}

func (a *Transaction) BeforeCreate(tx *gorm.DB) error {
	// go (nano) vs postgres (milli)
	a.CreatedAt = time.Now().Round(time.Millisecond)
	a.UpdatedAt = a.CreatedAt
	return nil
}

func (a *Transaction) AfterCreate(tx *gorm.DB) error {
	// cache user
	if err := a.Cache(); err != nil {
		return err
	}
	return nil
}

func (a *Transaction) BeforeUpdate(tx *gorm.DB) error {
	// go (nano) vs postgres (milli)
	a.UpdatedAt = time.Now().Round(time.Millisecond)
	return nil
}

// Updating data in same transaction
func (a *Transaction) AfterUpdate(tx *gorm.DB) error {
	// cache user
	if err := a.Cache(); err != nil {
		return err
	}
	return nil
}

func (a *Transaction) BeforeDelete(tx *gorm.DB) error {
	return nil
}

func (a *Transaction) AfterDelete(tx *gorm.DB) error {
	// rm cache user
	if err := a.DelCache(); err != nil {
		return err
	}
	return nil
}
