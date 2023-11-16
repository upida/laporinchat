package models

import (
	"gorm.io/gorm"
)

type ReportStatus struct {
	gorm.Model
	ReportID    uint   `json:"report_id"`
	Status      string `gorm:"text;size:255;not null"`
	Description string `gorm:"text;not null"`
	// Relation
	Report Report `gorm:"foreignkey:ReportID;references:ID" json:"report"`
}

func (report Report) GetDetailStatus() ([]ReportStatus, error) {
	var statuses []ReportStatus

	if err := DB.Where("report_id = ?", report.ID).Find(&statuses).Error; err != nil {
		return nil, err
	}

	return statuses, nil
}
