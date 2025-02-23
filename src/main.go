package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
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
	rtr.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./assets/images/crawler-logo.png") // Correct path to your image
		return
	})
	rtr.GET("/library", Auth.IsAuthenticated, func(c *gin.Context) {
		// Fetch image byte arrays
		data, err := Crawler.GetAllImagesForCustomer(Crawler.GetUser())
		if err != nil {
			log.Fatalf("Error fetching images: %v", err)
		}

		var goodImages [][]byte
		var base64Images []string // To store the Base64-encoded images

		for x := 0; x < len(data); x++ {
			img := data[x]
			// Decode the image to detect its type
			decodedImg, format, err := image.Decode(bytes.NewReader(img))
			if err != nil {
				// Log the error for invalid images and continue
				log.Printf("Failed to decode image at index %d: %v", x, err)
				continue
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

			// Set Content-Disposition to 'inline' so that the browser will display the image
			c.Header("Content-Type", contentType)
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

			// Append valid images to the goodImages slice
			goodImages = append(goodImages, buf.Bytes())

			// Convert image to Base64 for passing to the template
			base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())
			base64Images = append(base64Images, base64Img)
		}

		// Pass the encoded images to the template
		c.HTML(http.StatusOK, "library/libraryPage.gohtml", gin.H{
			"imgs": base64Images, // Use base64-encoded images
		})
	})

	rtr.GET("settings", Auth.IsAuthenticated, func(c *gin.Context) {
		c.HTML(http.StatusOK, "home/settings.gohtml", gin.H{})
	})
	rtr.GET("upload", Auth.IsAuthenticated, func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload/uploadPage.gohtml", gin.H{})
	})
	rtr.GET("imgRefGet", Auth.IsAuthenticated, func(c *gin.Context) {
		listNum := c.DefaultQuery("list", "0")
		num, err := strconv.Atoi(listNum)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Num"})
			return
		}
		img, err := Crawler.GetImageForCustomer(Crawler.GetUser(), int64(num))
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
	rtr.GET("/loadAlert", Auth.IsAuthenticated, func(c *gin.Context) {
		listNum := c.DefaultQuery("list", "0")
		num, err := strconv.Atoi(listNum)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Num"})
			return
		}

		data, err := Crawler.GetAlertTimestamp(Crawler.GetUser(), int64(num))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		// Return just a part of the page (unselect)
		c.HTML(200, "home/alert.gohtml", gin.H{
			"index":   num,
			"website": data[0],
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

	rtr.POST("/uploadFile", Auth.IsAuthenticated, func(c *gin.Context) {
		// Get the file from the form input
		file, err := c.FormFile("upload")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad file uploaded" + err.Error()})
			return
		}
		if file == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded" + err.Error()})
			return
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only image files are allowed." + err.Error()})
			return
		}
		byteFile, err := file.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad file uploaded" + err.Error()})
			return
		}
		data, err := io.ReadAll(byteFile)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad file uploaded" + err.Error()})
			return
		}
		Crawler.AddImageCustomer(data, Crawler.GetUser())
		c.HTML(http.StatusOK, "upload/uploadApprove.gohtml", gin.H{
			"file": file.Filename,
		})
	})
	rtr.GET("/uploadButton", Auth.IsAuthenticated, func(c *gin.Context) {
		// Return just a part of the page (upload button)
		c.HTML(http.StatusOK, "upload/uploadButton.gohtml", gin.H{})
	})
	rtr.Run(":8080")
}
