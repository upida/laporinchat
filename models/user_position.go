package models

import (
	"log"

	"gorm.io/gorm"
)

type UserPosition struct {
	gorm.Model
	UserID   uint64 `gorm:"text;size:255;not null;unique" json:"user_id"`
	Position string `gorm:"text;size:255;not null"`
}

func (user User) GetPosition() (UserPosition, error) {
	var position UserPosition

	err := DB.Where("user_id = ?", user.UserID).First(&position).Error
	if err != nil {
		return UserPosition{}, err
	}

	return position, nil
}

func (user User) SetPosition(name string) (UserPosition, error) {
	var err error
	var position UserPosition

	if current_position, err := user.GetPosition(); err == nil {
		position = current_position
		position.Position = name
		err = DB.Updates(&position).Error
	} else {
		log.Printf("Tidak punya posisi %s", err.Error())
		position.UserID = user.UserID
		position.Position = name
		err = DB.Create(&position).Error
	}

	if err != nil {
		return UserPosition{}, err
	}

	return position, nil
}
