package main

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/RedactedDog/crawler/src/Auth"
	"github.com/RedactedDog/crawler/src/Crawler"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}
	//start craweler thread
	go Crawler.Start()
	go Crawler.CheckerThread()
	authDef, err := Auth.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	rtr := Auth.NewRouter(authDef)
	rtr.LoadHTMLGlob("templates/**/*")
	rtr.StaticFS("/assets", http.Dir("assets"))
	rtr.GET("dashboard", Auth.IsAuthenticated, func(c *gin.Context) {
		c.HTML(http.StatusOK, "home/dashboard.gohtml", gin.H{
			"name": Crawler.GetUser(),
		})
	})
	rtr.GET("settings", Auth.IsAuthenticated, func(c *gin.Context) {
		c.HTML(http.StatusOK, "home/settings.gohtml", gin.H{})
	})
	rtr.GET("upload", Auth.IsAuthenticated, func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload/upload.gohtml", gin.H{})
	})

	rtr.GET("imgAlert", Auth.IsAuthenticated, func(c *gin.Context) {
		c.File("./shirt.jpg")
	})
	rtr.GET("imgRef", Auth.IsAuthenticated, func(c *gin.Context) {
		c.File("./proxy-image.jpg")
	})
	rtr.GET("/load-alert<>", Auth.IsAuthenticated, func(c *gin.Context) {
		// Return just a part of the page (template alert)
		c.HTML(200, "home/alert.gohtml", gin.H{
			"website": "AMAZON",
		})
	})
	rtr.GET("/unload-alert", Auth.IsAuthenticated, func(c *gin.Context) {
		// Return just a part of the page (unselect)
		c.HTML(200, "home/unselect.gohtml", gin.H{})
	})
	rtr.GET("/alert-list", Auth.IsAuthenticated, func(c *gin.Context) {
		data, err := Crawler.GetAlertTimestamps(Crawler.GetUser())
		if err != nil {
			log.Fatalf("There was an error dumbass: %v", err)
		}
		// Return just a part of the page (unselect)
		c.HTML(200, "home/alert-list.gohtml", gin.H{
			"alerts": data,
		})
	})

	rtr.GET("/upload-file", Auth.IsAuthenticated, func(c *gin.Context) {
		// Get the file from the form input
		file, _ := c.FormFile("file")
		if file == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only image files are allowed."})
			return
		}
	})
	// rtr.GET("user", Auth.IsAuthenticated, func(ctx *gin.Context) {

	// })
	rtr.Run(":8080")
}
