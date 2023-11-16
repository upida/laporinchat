package models

import (
	"gorm.io/gorm"
)

type UserVariable struct {
	gorm.Model
	UserID    uint64 `gorm:"text;size:255;not null" json:"user_id"`
	Name      string `gorm:"text;size:255;not null"`
	Value     string `gorm:"text"`
	Permanent bool   `gorm:"default:false"`
}

func (user User) GetVariable(name string) string {
	var variable UserVariable

	err := DB.Where("user_id = ?", user.UserID).Where("name = ?", name).First(&variable).Error
	if err != nil {
		return ""
	}

	return variable.Value
}

func (user User) SetVariable(name string, value string) (string, bool) {
	var err error
	var variable UserVariable

	err = DB.Where("user_id = ?", user.UserID).Where("name = ?", name).First(&variable).Error
	if err == nil {
		variable.Value = value
		err = DB.Updates(&variable).Error
	} else {
		variable.UserID = user.UserID
		variable.Name = name
		variable.Value = value
		err = DB.Create(&variable).Error
	}

	if err != nil {
		return "", false
	}

	return variable.Value, true
}
