package Crawler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const IMGNAME = "IMG"

func client() *redis.Client {
	c := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password set
		DB:       0,  // Use default DB
		Protocol: 2,  // Connection protocol
	})
	return c
}

func AddImageAmazon(url string) error {
	log.Println("Added Amazon URL: " + url)
	c := client()
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel() // Ensure context is canceled after use
	data, err := downloadImage(url)
	if err != nil {
		return err
	}
	//add with a exp of 2 week
	err = c.LPush(ctx, "Amazon", data, time.Hour*24*14).Err()
	if err != nil {
		return fmt.Errorf("failed to store image: %w", err)
	}
	return nil
}
func TestAddImageAmazon(data []byte) error {
	c := client()
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel() // Ensure context is canceled after use
	//add with a exp of 2 week
	err := c.LPush(ctx, "Amazon", data, time.Hour*24*14).Err()
	if err != nil {
		return fmt.Errorf("failed to store image: %w", err)
	}
	return nil
}

func GetAllImagesAmazon() ([][]byte, error) {
	c := client()
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get all the images from the "Amazon" list
	data, err := c.LRange(ctx, "Amazon", 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve images: %w", err)
	}

	var images [][]byte
	for _, item := range data {
		// Convert the Redis list data (string) back to a byte slice
		imgData := []byte(item)
		images = append(images, imgData)
	}

	return images, nil
}

func AddImageCustomer(image []byte, customer string) error {
	c := client()   // Ensure this function properly initializes Redis connection
	defer c.Close() // Close the connection when done (if not using a persistent client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Ensure context is canceled after use

	// Store image in a list (multiple images per customer)
	err := c.LPush(ctx, fmt.Sprintf("%s:%s", customer, IMGNAME), image).Err()
	if err != nil {
		return fmt.Errorf("failed to store image: %w", err)
	}

	return nil
}

func GetImageForCustomer(customer string, what int64) ([]byte, error) {
	c := client() // Initialize Redis client
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get all images from the list
	image, err := c.LIndex(ctx, customer+":IMG", what).Result()
	if err != nil {
		return nil, err
	}

	return []byte(image), nil
}

func GetAllImagesForCustomer(customer string) ([][]byte, error) {
	c := client()
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Retrieve all images for the customer (use a large range to get all elements)
	data, err := c.LRange(ctx, fmt.Sprintf("%s:%s", customer, IMGNAME), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve images: %w", err)
	}

	// Convert the string slices to byte slices
	var images [][]byte
	for _, item := range data {
		images = append(images, []byte(item))
	}

	return images, nil
}

func getCurrentUnixTime() int64 {
	return time.Now().Unix()
}

func formatUnixTime(unixTime int64) string {
	date := time.Unix(unixTime, 0).Format("2006-01-02 15:04:05")
	return date
}

func AlertUser(customer string, image, customerimage []byte, source string) error {
	c := client()   // Ensure this function properly initializes Redis connection
	defer c.Close() // Close the connection when done (if not using a persistent client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Ensure context is canceled after use
	len, err := c.SAdd(ctx, customer+"imagealertsdup", image).Result()
	if err != nil {
		return err
	}
	if len == 0 {
		return errors.New("image Already Added")
	}
	err = c.LPush(ctx, customer+"imagealerts", image).Err()
	if err != nil {
		return err
	}
	err = c.LPush(ctx, customer+"imagealertsconflict", customerimage).Err()
	if err != nil {
		return err
	}
	err = c.LPush(ctx, customer+"alert", source+","+formatUnixTime(getCurrentUnixTime())).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetAlertImage(customer string, what int64) ([]byte, error) {
	c := client() // Initialize Redis client
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get all images from the list
	image, err := c.LIndex(ctx, customer+"imagealerts", what).Result()
	if err != nil {
		return nil, err
	}

	return []byte(image), nil
}

func DeleteImageConflict(customer string, what int64) error {
	c := client() // Initialize Redis client
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get all images from the list
	image, err := c.LIndex(ctx, customer+"imagealertsconflict", what).Result()
	if err != nil {
		return err
	}
	err = c.LRem(ctx, customer+"imagealertsconflict", 1, image).Err()
	if err != nil {
		return err
	}
	// Get all images from the list
	image, err = c.LIndex(ctx, customer+"imagealerts", what).Result()
	if err != nil {
		return err
	}
	err = c.LRem(ctx, customer+"imagealerts", 1, image).Err()
	if err != nil {
		return err
	}
	strData, err := c.LIndex(ctx, customer+"alert", what).Result()
	if err != nil {
		return err
	}
	err = c.LRem(ctx, customer+"alert", 1, strData).Err()
	if err != nil {
		return err
	}
	return nil
}

func DeleteImageCustomer(customer string, what int64) error {
	c := client() // Initialize Redis client
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get all images from the list
	image, err := c.LIndex(ctx, fmt.Sprintf("%s:%s", customer, IMGNAME), what).Result()
	if err != nil {
		return err
	}
	err = c.LRem(ctx, fmt.Sprintf("%s:%s", customer, IMGNAME), 1, image).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetAlertConflict(customer string, what int64) ([]byte, error) {
	c := client() // Initialize Redis client
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get all images from the list
	image, err := c.LIndex(ctx, customer+"imagealertsconflict", what).Result()
	if err != nil {
		return nil, err
	}

	return []byte(image), nil
}

type AlertData struct {
	Time    string
	Website string
}

func GetAlertTimestamps(customer string) ([]AlertData, error) {
	c := client() // Initialize Redis client
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Retrieve all alert entries
	alerts, err := c.LRange(ctx, customer+"alert", 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve alerts: %w", err)
	}
	dataList := make([]AlertData, len(alerts))
	for x := 0; x < len(alerts); x++ {
		split := strings.Split(alerts[x], ",")
		if len(split) >= 2 {
			dataList[x] = AlertData{
				Website: split[0],
				Time:    split[1],
			}
		} else {
			// Handle case where split doesn't produce expected result
			return nil, fmt.Errorf("invalid alert format: %v", alerts[x])
		}
	}

	return dataList, nil
}

func GetAlertTimestamp(customer string, index int64) ([]string, error) {
	c := client() // Initialize Redis client
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch a single alert at the given index
	alert, err := c.LIndex(ctx, customer+"alert", index).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve alert: %w", err)
	}

	return strings.Split(alert, ","), nil
}
