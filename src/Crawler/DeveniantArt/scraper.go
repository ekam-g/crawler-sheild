package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ArtPost struct {
	Title     string    `json:"title"`
	Artist    string    `json:"artist"`
	URL       string    `json:"url"`
	ImageURL  string    `json:"image_url"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	searchTerm := "helluva boss"
	fmt.Printf("Searching for: %s\n", searchTerm)

	posts, err := scrapeDeviantArt(searchTerm, 50)
	if err != nil {
		log.Fatal(err)
	}

	// Print results
	fmt.Printf("\nFound %d posts:\n", len(posts))
	for i, post := range posts {
		fmt.Printf("\n=== Post %d ===\n", i+1)
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("Artist: %s\n", post.Artist)
		fmt.Printf("URL: %s\n", post.URL)
		fmt.Printf("Image URL: %s\n", post.ImageURL)
		fmt.Printf("==================\n")
	}

	// Save to JSON file
	filename := "helluva_boss_posts.json"
	err = saveToJSON(posts, filename)
	if err != nil {
		log.Printf("Error saving to JSON: %v", err)
	} else {
		fmt.Printf("\nSaved results to %s\n", filename)
	}
}

func scrapeDeviantArt(searchTerm string, limit int) ([]ArtPost, error) {
	baseURL := fmt.Sprintf("https://www.deviantart.com/search?q=%s", strings.ReplaceAll(searchTerm, " ", "+"))

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add headers to mimic browser request
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	fmt.Println("Sending request to DeviantArt...")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	fmt.Println("Parsing response...")
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	var posts []ArtPost

	// Updated selectors based on the provided HTML structure
	doc.Find("div._3Y0hT").Each(func(i int, s *goquery.Selection) {
		if i >= limit {
			return
		}

		// Get the link and title
		link := s.Find("a").First()
		url, _ := link.Attr("href")

		// Get the image URL
		img := s.Find("img").First()
		imageURL, _ := img.Attr("src")
		title := img.AttrOr("alt", "")

		// Get the artist name from the user-link class
		artist := s.Find("a.user-link._2yXGz span._2EfV7").Text()

		// Clean up the data
		title = strings.TrimSpace(title)
		artist = strings.TrimSpace(artist)

		post := ArtPost{
			Title:     title,
			Artist:    artist,
			URL:       url,
			ImageURL:  imageURL,
			Timestamp: time.Now(),
		}

		if post.Title != "" || post.URL != "" {
			posts = append(posts, post)
			fmt.Printf("Found post: %s by %s\n", post.Title, post.Artist)
		}
	})

	if len(posts) == 0 {
		fmt.Println("Warning: No posts found. HTML structure might have changed.")
		// Print a sample of the HTML for debugging
		html, _ := doc.Html()
		fmt.Printf("First 500 characters of HTML:\n%s\n", html[:min(len(html), 500)])
	}

	return posts, nil
}

func saveToJSON(posts []ArtPost, filename string) error {
	file, err := json.MarshalIndent(posts, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	err = os.WriteFile(filename, file, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
