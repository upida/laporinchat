package models

import (
	"encoding/json"
	"log"
	"strconv"

	"gorm.io/gorm"
)

type Report struct {
	gorm.Model
	UserID      uint64 `gorm:"text;size:255;not null" json:"user_id"`
	Title       string `gorm:"text;size:255;not null"`
	Description string `gorm:"text;not null"`
	Media       string `gorm:"text"`
	LastStatus  string `gorm:"text;size:255;not null"`
}

func (report Report) SetReport() (Report, error) {
	var err error

	logJson, _ := json.Marshal(report)
	log.Printf("SetReport ~ START: %s", logJson)

	err = DB.Create(&report).Error
	if err != nil {
		return Report{}, err
	}

	report, _, err = report.SetStatus("request")
	if err != nil {
		return Report{}, err
	}

	var admin []User
	admin, err = GetAdmin()
	if err != nil {
		log.Printf("SetReport ~ ERROR: %s", err.Error())
	} else if len(admin) > 0 {

		response := "*NEW REPORT* Report ID: " + strconv.FormatUint(uint64(report.ID), 10) + "\n" + "Status: " + report.LastStatus + "\n" + "Title: " + report.Title + "\n\n" + report.Description

		for _, user := range admin {
			if report.Media != "" {
				user.SendMessage(true, "Image URL: "+report.Media+"\n\n"+response)
				// user.SendPhoto(true, report.Media, response)
			} else {
				user.SendMessage(true, response)
			}
		}
	}

	return report, nil
}

func (report Report) SetStatus(status string, description ...string) (Report, ReportStatus, error) {
	var err error
	var reportStatus ReportStatus

	reportStatus.ReportID = report.ID
	reportStatus.Status = status

	if len(description) > 0 {
		reportStatus.Description = description[0]
	}

	err = DB.Create(&reportStatus).Error
	if err != nil {
		return Report{}, ReportStatus{}, err
	}

	report.LastStatus = status
	err = DB.Updates(&report).Error
	if err != nil {
		return Report{}, ReportStatus{}, err
	}

	user, err := GetUserByID(report.UserID)
	if err != nil {
		return Report{}, ReportStatus{}, err
	}

	if report.LastStatus != "request" {
		message := "*UPDATE* Report ID " + strconv.FormatUint(uint64(report.ID), 10) + "\n" +
			"Title: " + report.Title + "\n" +
			"Status: " + report.LastStatus + "\n\n"

		if reportStatus.Description != "" {
			message += "Status Description:\n" + reportStatus.Description
		}
		user.SendMessage(false, message)
	}

	return report, reportStatus, nil
}

type GetReportWhere struct {
	ID     *uint
	UserID *uint64
	Status *string
}

func GetReport(where GetReportWhere) ([]Report, error) {
	var reports []Report

	query := DB

	if where.ID != nil {
		query = query.Where("id = ?", *where.ID)
	}

	if where.UserID != nil {
		query = query.Where("user_id = ?", *where.UserID)
	}

	if where.Status != nil {
		query = query.Where("last_status = ?", *where.Status)
	}

	if err := query.Find(&reports).Error; err != nil {
		return nil, err
	}

	return reports, nil
}
