package webhook

import (
	"laporinchat/models"
	"net/http"
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

	var position models.UserPosition
	position, err = user.GetPosition()
	if err != nil {
		position, err = user.SetPosition("start")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error})
		}
	}

	if request.Message.Text == "/cancel" {
		user.SetVariable("Report ID", "")
		user.SetVariable("Report Status", "")
		if user.Admin {
			showMenu(user)
			user.SetPosition("verified")
		} else {
			user.SendMessage(true, "Please input your admin code")
			user.SetPosition("start")
		}
	} else if !user.Admin && position.Position == "start" {
		user.SendMessage(true, "Please input your admin code")
		user.SetPosition("verify")
	} else if !user.Admin && position.Position == "verify" {
		check := user.VerifyAdminRequest(request.Message.Text)
		if check {
			user.SendMessage(true, "Congratulations! You will receive notifications from user reports")
			user.SendMessage(true, "Menu:\n"+
				"/report_request : Show request reports\n"+
				"/report_processed : Show processed reports\n"+
				"/report_redirected : Show redirected reports\n"+
				"/report_declined : Show declined reports\n"+
				"/report_completed : Show completed reports")
			user.SetPosition("verified")
		} else {
			user.SendMessage(true, "Sorry! Your code does not match any code. Please send the correct code!")
		}
	} else if user.Admin && position.Position == "verified" {
		if request.Message.Text == "/report_request" {
			showReport(user, "request")
		} else if request.Message.Text == "/report_processed" {
			showReport(user, "processed")
		} else if request.Message.Text == "/report_redirected" {
			showReport(user, "redirected")
		} else if request.Message.Text == "/report_declined" {
			showReport(user, "declined")
		} else if request.Message.Text == "/report_completed" {
			showReport(user, "completed")
		} else {
			ID, err := strconv.ParseUint(request.Message.Text, 10, 64)
			if err == nil {
				ID32 := uint(ID)

				where := models.GetReportWhere{
					ID: &ID32,
				}
				reports, _ := models.GetReport(where)
				if len(reports) > 0 {
					user.SetVariable("Report ID", strconv.FormatUint(uint64(reports[0].ID), 10))
					user.SetPosition("update_status")
					updateStatusPosition(user)
				} else {
					user.SendMessage(true, "Report not found")
					showMenu(user)
				}
			} else {
				showMenu(user)
			}
		}
	} else if user.Admin && position.Position == "update_status" && user.GetVariable("Report ID") != "" {
		updateStatusPosition(user)
	} else if user.Admin && position.Position == "update_status_set_status" && user.GetVariable("Report ID") != "" {
		updateStatusSetStatusPosition(user, request.Message.Text)
	} else if user.Admin && position.Position == "update_status_set_description" && user.GetVariable("Report ID") != "" && user.GetVariable("Report Status") != "" {
		updateStatusSetDescriptionPosition(user, request.Message.Text)
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "data": request})
}

func showReport(user models.User, status string) {
	var button []models.Button
	button = append(button, models.Button{
		Text:         "/cancel",
		CallbackData: "/cancel",
	})

	request := models.GetReportWhere{
		Status: &status,
	}

	reports, _ := models.GetReport(request)

	if len(reports) > 0 {
		for _, report := range reports {
			response := "Report ID: *" + strconv.FormatUint(uint64(report.ID), 10) + "*\n" + "Status: " + report.LastStatus + "\n" + "Title: " + report.Title + "\n\n" + report.Description
			if report.Media != "" {
				user.SendMessage(true, "Image URL: "+report.Media+"\n\n"+response)
				// user.SendPhoto(true, report.Media, response)
			} else {
				user.SendMessage(true, response)
			}
		}
		user.SendMessage(true, "Enter Report ID to view details", button)
	} else {
		user.SendMessage(true, "There is no "+status+" reports!", button)
		return
	}

	return
}

func showMenu(user models.User) {
	user.SendMessage(true, "You will receive notifications from user reports")
	user.SendMessage(true, "Menu:\n"+
		"/report_request : Show request reports\n"+
		"/report_processed : Show processed reports\n"+
		"/report_redirected : Show redirected reports\n"+
		"/report_declined : Show declined reports\n"+
		"/report_completed : Show completed reports")
}

func updateStatusPosition(user models.User) {
	var button []models.Button
	button = append(button, models.Button{
		Text:         "/cancel",
		CallbackData: "/cancel",
	})

	ReportID := user.GetVariable("Report ID")

	ID, err := strconv.ParseUint(ReportID, 10, 64)
	if err == nil {
		ID32 := uint(ID)

		where := models.GetReportWhere{
			ID: &ID32,
		}
		reports, err := models.GetReport(where)
		if err != nil {
			user.SendMessage(true, "Your Report ID not found. Enter the correct Report ID to update status", button)
			user.SetVariable("Report ID", "")
			user.SetPosition("verified")
		} else if len(reports) > 0 {
			user.SendMessage(true, "Please enter new status", button)
			user.SetPosition("update_status_set_status")
		}
	}
}

func updateStatusSetStatusPosition(user models.User, status string) {
	var button []models.Button
	button = append(button, models.Button{
		Text:         "/cancel",
		CallbackData: "/cancel",
	})

	ReportID := user.GetVariable("Report ID")

	ID, err := strconv.ParseUint(ReportID, 10, 64)
	if err == nil {
		ID32 := uint(ID)

		where := models.GetReportWhere{
			ID: &ID32,
		}
		reports, err := models.GetReport(where)
		if err != nil {
			user.SendMessage(true, "Your Report ID not found. Enter the correct Report ID to update status", button)
			user.SetVariable("Report ID", "")
			user.SetPosition("verified")
		} else if len(reports) > 0 {
			user.SendMessage(true, "Please enter new description", button)
			user.SetVariable("Report Status", status)
			user.SetPosition("update_status_set_description")
		}
	}
}

func updateStatusSetDescriptionPosition(user models.User, description string) {
	var button []models.Button
	button = append(button, models.Button{
		Text:         "/cancel",
		CallbackData: "/cancel",
	})

	ReportID := user.GetVariable("Report ID")
	ReportStatus := user.GetVariable("Report Status")

	ID, err := strconv.ParseUint(ReportID, 10, 64)
	if err == nil {
		ID32 := uint(ID)

		where := models.GetReportWhere{
			ID: &ID32,
		}
		reports, err := models.GetReport(where)
		if err != nil {
			user.SendMessage(true, "Your Report ID not found. Enter the correct Report ID to update status", button)
			user.SetVariable("Report ID", "")
			user.SetPosition("verified")
		} else if len(reports) > 0 {
			reports[0].SetStatus(ReportStatus, description)
			user.SendMessage(true, "Thanks for your update! I will notify user of their reports")
			user.SetVariable("Report ID", "")
			user.SetVariable("Report Status", "")
			user.SetPosition("verified")
		}
	}
}
