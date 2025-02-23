package Crawler

import (
	"math/rand"
	"time"

	Deveniantart "github.com/RedactedDog/crawler/src/Crawler/DeveniantArt"
)

func Start() {
	apparelTypes := [12]string{
		"Graphic t-shirts, tank tops, long-sleeve shirts",
		"Hooded sweatshirts, pullovers",
		"Baseball caps, beanies, fedoras",
		"Ankle socks, crew socks, knee-high socks",
		"Jeans, joggers, sweatpants, athletic pants",
		"T-shirts dresses, sundresses, evening gowns",
		"Leather jackets, denim jackets, bomber jackets",
		"Silk scarves, woolen scarves, bandanas",
		"Handbags, backpacks, tote bags, duffel bags",
		"Ceramic mugs, travel mugs, insulated tumblers",
		"Bath towels, beach towels, hand towels",
		"Golf gloves, work gloves, touchscreen gloves",
	}
	apparelNum := -1
	for {
		apparelNum += 1
		if apparelNum == len(apparelTypes) {
			apparelNum = 0
		}
		rand.Seed(time.Now().UnixNano())
		num := rand.Intn(11) + 5
		// sleep 5 - 15 Hours
		time.Sleep(time.Duration(num) * time.Hour)

		// cloths, err := Amazon.FindTop50(apparelTypes[apparelNum])
		cloths, err := Deveniantart.Scrape()
		if err != nil {
			continue
		}
		for i := 0; i < len(cloths); i += 1 {
			//todo add to database
			go AddImageAmazon(cloths[i].ImageURL)
		}

	}
}
