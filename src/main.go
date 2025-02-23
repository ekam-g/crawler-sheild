package main

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
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
		c.HTML(http.StatusOK, "upload/uploadPage.gohtml", gin.H{})
	})

	rtr.GET("imgRef", Auth.IsAuthenticated, func(c *gin.Context) {
		listNum := c.DefaultQuery("list", "0")
		num, err := strconv.Atoi(listNum)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Num"})
			return
		}
		img, err := Crawler.GetAlertConflict(Crawler.GetUser(), int64(num))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		// Decode the image to detect its type
		decodedImg, format, err := image.Decode(bytes.NewReader(img))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode image"})
			return
		}

		// Set the correct Content-Type based on the image format
		var contentType string
		switch format {
		case "jpeg":
			contentType = "image/jpeg"
		case "png":
			contentType = "image/png"
		case "gif":
			contentType = "image/gif"
		default:
			contentType = "application/octet-stream"
		}

		// Set the content type header to the detected type
		c.Header("Content-Type", contentType)

		// Set Content-Disposition to 'inline' so that the browser will display the image
		c.Header("Content-Disposition", "inline; filename=image."+format)

		// Encode the image back to the appropriate format and send it as the response
		var buf bytes.Buffer
		switch format {
		case "jpeg":
			err = jpeg.Encode(&buf, decodedImg, nil)
		case "png":
			err = png.Encode(&buf, decodedImg)
		case "gif":
			// If it's a GIF, encode accordingly (you can add more formats as needed)
			err = gif.Encode(&buf, decodedImg, nil)
		default:
			err = nil
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode image"})
			return
		}

		// Return the image data
		c.Data(http.StatusOK, contentType, buf.Bytes())
	})

	rtr.GET("imgAlert", Auth.IsAuthenticated, func(c *gin.Context) {
		listNum := c.DefaultQuery("list", "0")
		num, err := strconv.Atoi(listNum)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Num"})
			return
		}
		img, err := Crawler.GetAlertImage(Crawler.GetUser(), int64(num))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		// Decode the image to detect its type
		decodedImg, format, err := image.Decode(bytes.NewReader(img))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode image"})
			return
		}

		// Set the correct Content-Type based on the image format
		var contentType string
		switch format {
		case "jpeg":
			contentType = "image/jpeg"
		case "png":
			contentType = "image/png"
		case "gif":
			contentType = "image/gif"
		default:
			contentType = "application/octet-stream"
		}

		// Set the content type header to the detected type
		c.Header("Content-Type", contentType)

		// Set Content-Disposition to 'inline' so that the browser will display the image
		c.Header("Content-Disposition", "inline; filename=image."+format)

		// Encode the image back to the appropriate format and send it as the response
		var buf bytes.Buffer
		switch format {
		case "jpeg":
			err = jpeg.Encode(&buf, decodedImg, nil)
		case "png":
			err = png.Encode(&buf, decodedImg)
		case "gif":
			// If it's a GIF, encode accordingly (you can add more formats as needed)
			err = gif.Encode(&buf, decodedImg, nil)
		default:
			err = nil
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode image"})
			return
		}

		// Return the image data
		c.Data(http.StatusOK, contentType, buf.Bytes())
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
