package main

import (
	"log"
	"os"

	"github.com/RedactedDog/crawler/src/Crawler"
)

func main() {
	//data, err := Amazon.FindTop50("shirts")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//for x := 0; x < len(data); x += 1 {
	//	fmt.Println(data[x].ImageURL)
	//}
	addFileAmazon()
	// addFileUser()
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
