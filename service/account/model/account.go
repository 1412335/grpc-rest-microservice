package model

import (
	"encoding/json"
	"fmt"
	"time"

	pb "account/api"
	errorSrv "account/error"

	"github.com/1412335/grpc-rest-microservice/pkg/cache"
	"github.com/1412335/grpc-rest-microservice/pkg/errors"

	"github.com/microcosm-cc/bluemonday"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

type Account struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id" validate:"nonzero"`
	Name      string    `json:"name" validate:"max=100"`
	Bank      string    `json:"bank" validate:"nonzero"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *Account) Transform2GRPC() *pb.Account {
	acc := &pb.Account{
		Id:        a.ID,
		Name:      a.Name,
		UserId:    a.UserID,
		Balance:   a.Balance,
		CreatedAt: timestamppb.New(a.CreatedAt),
	}
	//
	if bank, ok := pb.Bank_value[a.Bank]; ok {
		acc.Bank = pb.Bank(bank)
	}
	return acc
}

func (a *Account) UpdateFromGRPC(acc *pb.Account) {
	a.UserID = acc.GetUserId()
	a.Name = acc.GetName()
	a.Bank = acc.GetBank().String()
}

func (a *Account) GetCache() error {
	var bytes []byte
	if err := cache.Get(a.ID, &bytes); err == cache.ErrCacheNotAvailable {
		return nil
	} else if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, a); err != nil {
		return err
	}
	return nil
}

func (a *Account) Cache() error {
	if bytes, err := json.Marshal(a); err != nil {
		return err
	} else if err := cache.Set(a.ID, string(bytes)); err == cache.ErrCacheNotAvailable {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

func (a *Account) DelCache() error {
	if err := cache.Delete(a.ID); err == cache.ErrCacheNotAvailable {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

func (a *Account) sanitize() {
	p := bluemonday.UGCPolicy()
	a.Name = p.Sanitize(a.Name)
}

func (a *Account) Validate() error {
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

func (a *Account) BeforeCreate(tx *gorm.DB) error {
	if a.Name == "" {
		a.Name = fmt.Sprintf("%s.%s", a.UserID, a.Bank)
	}
	return nil
}

func (a *Account) AfterCreate(tx *gorm.DB) error {
	// cache user
	if err := a.Cache(); err != nil {
		return err
	}
	return nil
}

func (a *Account) BeforeUpdate(tx *gorm.DB) error {
	if a.Name == "" {
		a.Name = fmt.Sprintf("%s.%s", a.UserID, a.Bank)
	}
	return nil
}

// Updating data in same transaction
func (a *Account) AfterUpdate(tx *gorm.DB) error {
	// find user by id
	if e := tx.First(a).Error; e == gorm.ErrRecordNotFound {
		return errorSrv.ErrAccountNotFound
	} else if e != nil {
		return errorSrv.ErrConnectDB
	}
	// cache user
	if err := a.Cache(); err != nil {
		return err
	}
	return nil
}

func (a *Account) BeforeDelete(tx *gorm.DB) error {
	return nil
}

func (a *Account) AfterDelete(tx *gorm.DB) error {
	// rm cache user
	if err := a.DelCache(); err != nil {
		return err
	}
	return nil
}
