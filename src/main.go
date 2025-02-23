package main

import (
	"log"
	"net/http"

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
	rtr.GET("imgAlert", Auth.IsAuthenticated, func(c *gin.Context) {
		c.File("./shirt.jpg")
	})
	rtr.GET("imgRef", Auth.IsAuthenticated, func(c *gin.Context) {
		c.File("./proxy-image.jpg")
	})
	rtr.GET("/load-alert", Auth.IsAuthenticated, func(c *gin.Context) {
		// Return just a part of the page (template alert)
		c.HTML(200, "home/alert.gohtml", gin.H{
			"website": "AMAZON",
		})
	})
	rtr.GET("/unload-alert", Auth.IsAuthenticated, func(c *gin.Context) {
		// Return just a part of the page (unselect)
		c.HTML(200, "home/unselect.gohtml", gin.H{})
	})
	// rtr.GET("user", Auth.IsAuthenticated, func(ctx *gin.Context) {

	// })
	rtr.Run(":8080")
}
