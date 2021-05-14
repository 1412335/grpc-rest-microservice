package v3

import (
	"encoding/json"
	"strings"
	"time"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"
	"github.com/1412335/grpc-rest-microservice/pkg/errors"
	"github.com/1412335/grpc-rest-microservice/pkg/utils"
	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

type User struct {
	ID          string `json:"id"`
	Username    string `validate:"nonzero,regexp=^[a-zA-Z0-9_]*$"`
	Fullname    string `validate:"nonzero,max=100"`
	Active      bool
	Password    string `json:"-" validate:"min=8"`
	Email       string `gorm:"uniqueIndex" validate:"nonzero"`
	VerifyToken string
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (u *User) transform2GRPC() *api_v3.User {
	user := &api_v3.User{
		Id:          u.ID,
		Username:    u.Username,
		Fullname:    u.Fullname,
		Active:      u.Active,
		Email:       u.Email,
		VerifyToken: u.VerifyToken,
	}
	if role, ok := api_v3.Role_value[u.Role]; ok {
		user.Role = api_v3.Role(role)
	}
	return user
}

func (u *User) updateFromGRPC(user *api_v3.User) {
	u.Username = user.GetUsername()
	u.Fullname = user.GetFullname()
	u.Email = user.GetEmail()
	u.Password = user.GetPassword()
	// check gte admin
	if role, ok := api_v3.Role_value[u.Role]; ok && api_v3.Role(role) >= api_v3.Role_ADMIN {
		u.Role = user.GetRole().String()
		u.Active = user.GetActive()
	}
}

func (u *User) cache() error {
	if DefaultCache == nil {
		return nil
	}
	if bytes, err := json.Marshal(u); err != nil {
		return err
	} else if err := DefaultCache.Set(u.ID, string(bytes)); err != nil {
		return err
	}
	return nil
}

func (u *User) rmCache() error {
	if DefaultCache == nil {
		return nil
	}
	if err := DefaultCache.Delete(u.ID); err != nil {
		return err
	}
	return nil
}

func (u *User) hashPassword() error {
	// hash password
	hashedPassword, err := utils.GenHash(u.Password)
	if err != nil {
		// 	u.logger.For(ctx).Error("Hash password failed", zap.Error(err))
		return ErrHashPassword
	}
	u.Password = hashedPassword
	return nil
}

func (u *User) sanitize() {
	p := bluemonday.UGCPolicy()
	u.Username = p.Sanitize(u.Username)
	u.Fullname = p.Sanitize(u.Fullname)
}

func (u *User) validate() error {
	// sanitize fileds
	u.sanitize()
	// validate
	if e := validator.Validate(u); e != nil {
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

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if err := u.hashPassword(); err != nil {
		return err
	}
	u.Email = strings.ToLower(u.Email)
	return nil
}

func (u *User) AfterCreate(tx *gorm.DB) error {
	// cache user
	if err := u.cache(); err != nil {
		return err
	}
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if err := u.hashPassword(); err != nil {
		return err
	}
	u.Email = strings.ToLower(u.Email)
	return nil
}

// Updating data in same transaction
func (u *User) AfterUpdate(tx *gorm.DB) error {
	// cache user
	if err := u.cache(); err != nil {
		return err
	}
	return nil
}

func (u *User) BeforeDelete(tx *gorm.DB) error {
	return nil
}

func (u *User) AfterDelete(tx *gorm.DB) error {
	// rm cache user
	if err := u.rmCache(); err != nil {
		return err
	}
	return nil
}
