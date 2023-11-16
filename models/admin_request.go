package models

import (
	"time"

	helper "laporinchat/utils"

	"gorm.io/gorm"
)

type AdminRequest struct {
	gorm.Model
	Username      string     `gorm:"text;size:255;not null;unique"`
	Code          string     `gorm:"text;size:6;not null"`
	VerificatedAt *time.Time `json:"verificated_at"`
}

func SetAdminRequest(username string) (AdminRequest, error) {
	var err error
	var request AdminRequest

	request.Username = username

	code := helper.RandomString(6)

	request.Code = code

	err = DB.Create(&request).Error
	if err != nil {
		return AdminRequest{}, err
	}
	return request, nil
}

func (user User) VerifyAdminRequest(code string) bool {
	var request AdminRequest

	err := DB.Where("username = ?", user.Username).Where("verificated_at is null").First(&request).Error
	if err != nil {
		return false
	}

	if request.Code != code {
		return false
	}

	_, err = user.SetAdmin()
	if err != nil {
		return false
	}

	now := time.Now()
	request.VerificatedAt = &now

	return true
}
