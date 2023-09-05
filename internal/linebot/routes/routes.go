package routes

import (
	"apex_tournament_noti/internal/linebot/webhook"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	LineBotChannelSecret = os.Getenv("LINE_BOT_CHANNEL_SECRET")
	LineBotChannelToken  = os.Getenv("LINE_BOT_CHANNEL_TOKEN")
)

// Create a Middleware to check the signature
func signatureMiddleware(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
		})
	}

	signiture := c.Request.Header.Get("x-line-signature")
	decoded, err := base64.StdEncoding.DecodeString(signiture)
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
		})
	}
	hash := hmac.New(sha256.New, []byte(LineBotChannelSecret))
	hash.Write(body)

	// Compare decoded signature and `hash.Sum(nil)` by using `hmac.Equal`
	if !hmac.Equal(decoded, hash.Sum(nil)) {
		log.Print("Invalid signature (checked inside middleware)")
		c.AbortWithStatus(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid signature",
		})
	} else {
		log.Print("Valid signature (checked inside middleware)")
	}

	c.Next()
}

func SetupRoutes(r *gin.Engine, ch chan webhook.WebhookPayload) {
	r.POST("/webhook", signatureMiddleware, webhook.HandleWebhook(ch))
	r.POST("/test/webhook", webhook.TestHandleWebhook(ch))

	// Route for SSL certificate
	r.GET("/.well-known/pki-validation/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		fmt.Println("filename: ", filename)

		// Check if there is a file with the same name
		if _, err := os.Stat("/app/file_download/" + filename); err == nil {
			// Start downloading the file
			c.File("/app/file_download/" + filename)
			return
		} else {
			fmt.Println(err)
			c.JSON(http.StatusNotFound, gin.H{
				"message": "File not found",
			})
		}
	})
}
