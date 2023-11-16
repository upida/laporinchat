package models

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserID    uint64 `gorm:"text;size:255;not null;unique" json:"user_id"`
	Admin     bool   `gorm:"default:false"`
	Username  string `gorm:"text;size:255;not null;unique"`
	Name      string `gorm:"text;size:255;not null"`
	FirstName string `gorm:"text;size:255"`
	LastName  string `gorm:"text;size:255"`
	CardID    string `gorm:"text;size:255"`
}

func GetUserByID(userId uint64) (User, error) {

	var u User

	err := DB.Where("user_id = ?", userId).First(&u).Error
	if err != nil {
		return u, errors.New("User not found!")
	}

	return u, nil

}

func (user User) SetUser() (User, error) {
	var err error

	err = DB.Create(&user).Error
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func GetAdmin() ([]User, error) {

	var u []User

	err := DB.Where("admin = ?", true).Find(&u).Error
	if err != nil {
		return u, errors.New("Admin not found!")
	}

	return u, nil

}

func (user User) SetAdmin() (User, error) {
	var err error

	user.Admin = true

	err = DB.Updates(&user).Error
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (user User) SetCardId(cardId string) (User, error) {
	var err error

	user.CardID = cardId

	err = DB.Updates(&user).Error
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (user User) CreateReport(title string, description string, media ...string) (Report, error) {
	var err error

	var report Report
	report.UserID = uint64(user.ID)
	report.Title = strings.TrimSpace(title)
	report.Description = strings.TrimSpace(description)

	if len(media) > 0 {
		report.Media = strings.TrimSpace(media[0])
	}

	report, err = report.SetReport()
	if err != nil {
		return Report{}, err
	}

	return report, nil
}
