package v3

import (
	"time"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"

	"gorm.io/gorm"
)

type User struct {
	ID          string
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

func (u *User) BeforeCreate(tx *gorm.DB) error {
	return nil
}

func (u *User) AfterCreate(tx *gorm.DB) error {
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// Updating data in same transaction
func (u *User) AfterUpdate(tx *gorm.DB) error {
	return nil
}
