package middlewares

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
)

func RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": "Gagal membaca body"})
			c.Abort()
			return
		}

		var requestBody map[string]interface{}
		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			c.JSON(400, gin.H{"error": "Gagal melakukan unmarshal body"})
			c.Abort()
			return
		}

		responseJSON, err := json.Marshal(requestBody)
		if err != nil {
			c.JSON(500, gin.H{"error": "Gagal mengonversi ke JSON"})
			c.Abort()
			return
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		log.Printf("REQUEST:%s\n", responseJSON)
		c.Next()
	}
}
