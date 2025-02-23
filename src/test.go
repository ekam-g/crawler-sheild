package main

import (
	"log"
	"os"

	"github.com/RedactedDog/crawler/src/Crawler"
)

func main() {
	addFileAmazon()
	addFileUser()
}

func addFileAmazon() {
	file := "shirt.jpg"
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
		return
	}
	Crawler.TestAddImageAmazon(data)
}

func addFileUser() {
	file := "proxy-image.jpg"
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
		return
	}
	Crawler.AddImageCustomer(data, "Ekam")
}
