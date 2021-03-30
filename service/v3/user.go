package v3

import (
	"encoding/json"
	"strings"
	"time"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"
	"github.com/1412335/grpc-rest-microservice/pkg/utils"

	"gorm.io/gorm"
)

type User struct {
	ID          string `json:"id"`
	Username    string
	Fullname    string
	Active      bool
	Password    string `json:"-"`
	Email       string `gorm:"uniqueIndex"`
	VerifyToken string
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (u *User) sanitize() *api_v3.User {
	return &api_v3.User{
		Id:          u.ID,
		Username:    u.Username,
		Fullname:    u.Fullname,
		Active:      u.Active,
		Email:       u.Email,
		Role:        api_v3.Role(api_v3.Role_value[u.Role]),
		VerifyToken: u.VerifyToken,
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
	if u.Password != "" {
		return nil
	}
	// hash password
	hashedPassword, err := utils.GenHash(u.Password)
	if err != nil {
		// 	u.logger.For(ctx).Error("Hash password failed", zap.Error(err))
		return ErrHashPassword
	}
	u.Password = hashedPassword
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
