package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"laporinchat/models"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type InboundRequest struct {
	Message struct {
		Chat struct {
			ID uint64 `json:"id"`
		} `json:"chat"`
		From struct {
			ID        uint64 `json:"id"`
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		} `json:"from"`
		Text    string `json:"text"`
		Caption string `json:"caption"`
		Photo   []struct {
			FileID   string `json:"file_id"`
			FileName string `json:"file_name"`
		} `json:"photo"`
	} `json:"message"`

	CallbackQuery struct {
		Data    string `json:"data"`
		Message struct {
			Chat struct {
				ID uint64 `json:"id"`
			} `json:"chat"`
			From struct {
				ID        uint64 `json:"id"`
				Username  string `json:"username"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
			} `json:"from"`
			Text    string `json:"text"`
			Caption string `json:"caption"`
			Photo   []struct {
				FileID   string `json:"file_id"`
				FileName string `json:"file_name"`
			} `json:"photo"`
		} `json:"message"`
	} `json:"callback_query"`
}

func Inbound(c *gin.Context) {
	var err error
	var request InboundRequest

	if err = c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.CallbackQuery.Message.Chat.ID != 0 {
		request.Message = request.CallbackQuery.Message
		request.Message.Text = request.CallbackQuery.Data
	}

	var user models.User
	user, err = models.GetUserByID(uint64(request.Message.Chat.ID))

	if err != nil {
		user.UserID = request.Message.From.ID
		user.Username = request.Message.From.Username
		user.FirstName = request.Message.From.FirstName
		user.LastName = request.Message.From.LastName

		name := strings.Join([]string{request.Message.From.FirstName, request.Message.From.LastName}, " ")
		user.Name = strings.TrimSpace(name)

		user, _ = user.SetUser()
	}

	rule(c, user, request)

	c.JSON(http.StatusOK, gin.H{"status": true, "data": request})
}

func rule(c *gin.Context, user models.User, request InboundRequest) {
	var position models.UserPosition
	position, err := user.GetPosition()
	if err != nil {
		position, err = user.SetPosition("start")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error})
			c.Abort()
		}
	}

	var button []models.Button

	log.Printf("MESSAGE TEXT : %s", request.Message.Text)

	if request.Message.Text == "/cancel" {
		user.SetPosition("start")
		user.SetVariable("title", "")

		button = append(button, models.Button{
			Text:         "/your_reports",
			CallbackData: "/your_reports",
		})

		user.SendMessage(false, "Request cancelled!", button)
	} else if request.Message.Text == "/your_reports" || position.Position == "reports" {
		user.SetPosition("reports")

		button = append(button, models.Button{
			Text:         "/cancel",
			CallbackData: "/cancel",
		})

		where := models.GetReportWhere{
			UserID: &user.UserID,
		}

		if request.Message.Text != "/your_reports" {
			ID, err := strconv.ParseUint(request.Message.Text, 10, 64)
			if err != nil {
				user.SendMessage(false, "Your Report ID must be integer. Enter Report ID to view details", button)
			}
			ID32 := uint(ID)
			where.ID = &ID32
		}

		reports, _ := models.GetReport(where)

		if where.ID != nil {
			if len(reports) > 0 {
				statuses, _ := reports[0].GetDetailStatus()

				var response []string
				for _, status := range statuses {
					var text string
					text = "[" + status.CreatedAt.Format("2006-01-02 15:04:05") + "] "
					text += status.Status
					if status.Description != "" {
						text += "\nDescription: " + status.Description
					}
					response = append(response, text)
				}

				if len(response) > 0 {
					user.SendMessage(false, strings.Join(response, "\n"), button)
				} else {
					user.SendMessage(false, "Sorry! Failed to get status. Please try again!", button)
				}
			} else {
				user.SendMessage(false, "Sorry! Your Report ID not found. Please try again!", button)
			}
		} else {
			if len(reports) > 0 {
				for _, report := range reports {
					response := "Report ID: *" + strconv.FormatUint(uint64(report.ID), 10) + "*\n" + "Status: " + report.LastStatus + "\n" + "Title: " + report.Title + "\n\n" + report.Description
					if report.Media != "" {
						user.SendMessage(false, "Image URL: "+report.Media+"\n\n"+response)
						// user.SendPhoto(false, report.Media, response)
					} else {
						user.SendMessage(false, response)
					}
				}
				user.SendMessage(false, "Enter Report ID to view details", button)
			} else {
				user.SendMessage(false, "You don't have a report yet!", button)
				user.SetPosition("start")
			}
		}
	} else if position.Position == "start" {

		button = append(button, models.Button{
			Text:         "/cancel",
			CallbackData: "/cancel",
		})

		user.SendMessage(false, "Type your report title", button)
		user.SetPosition("title")
	} else if position.Position == "title" {

		button = append(button, models.Button{
			Text:         "/cancel",
			CallbackData: "/cancel",
		})

		user.SetVariable("title", request.Message.Text)
		user.SendMessage(false, "Type your report description."+
			"\n\nIf you have image please add report description as caption", button)
		user.SetPosition("description")
	} else if position.Position == "description" {

		button = append(button, models.Button{
			Text:         "/your_reports",
			CallbackData: "/your_reports",
		})

		var report models.Report

		report.UserID = user.UserID
		report.Title = user.GetVariable("title")

		if request.Message.Caption != "" {
			report.Description = request.Message.Caption
		} else {
			report.Description = request.Message.Text
		}

		if len(request.Message.Photo) > 0 {
			fileURL, _ := getFile(request.Message.Photo[0].FileID)
			report.Media = fileURL
		}

		report, _ = report.SetReport()
		user.SendMessage(false, "Thank you for your report! We will announce to you if there are any updates!", button)
		user.SetPosition("start")
		user.SetVariable("title", "")
	}
}

type fileResponse struct {
	OK     bool `json:"ok"`
	Result struct {
		FileID       string `json:"file_id"`
		FileUniqueID string `json:"file_unique_id"`
		FileSize     int    `json:"file_size"`
		FilePath     string `json:"file_path"`
	} `json:"result"`
}

func getFile(fileID string) (string, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s", token, fileID)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var fileResp fileResponse
	if err := json.Unmarshal(body, &fileResp); err != nil {
		return "", err
	}

	if !fileResp.OK {
		return "", fmt.Errorf("Telegram API error: %s", string(body))
	}

	baseURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/", token)
	path := fileResp.Result.FilePath

	fileURL := baseURL + path

	return fileURL, nil
}
