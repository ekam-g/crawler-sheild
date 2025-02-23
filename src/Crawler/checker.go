package Crawler

import (
	"log"
	"time"
)

func GetUser() string {
	return "Ekam"
}

func CheckCustomer(customer string) error {
	imgs, err := GetAllImagesForCustomer(customer)
	if err != nil {
		return err
	}
	amazonIMG, err := GetAllImagesAmazon()
	if err != nil {
		return err
	}
	for i := 0; i < len(imgs); i++ {
		go func() {
			img := imgs[i]
			for x := 0; x < len(amazonIMG); x += 1 {
				simlar, _ := orb(amazonIMG[x], img)
				if simlar {
					err := AlertUser(customer, amazonIMG[x], img, "Amazon")
					if err != nil {
						log.Println(err)
					}
				}
			}
		}()
	}
	return nil
}

func CheckerThread() {
	for {
		time.Sleep(10 * time.Second) // CHANGE LATER FOR DEMO
		user := GetUser()
		err := CheckCustomer(user)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Done One pass")
	}
}
