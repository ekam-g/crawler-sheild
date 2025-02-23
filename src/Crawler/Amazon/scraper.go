package Amazon

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Product represents a shirt listing
type Product struct {
	Title     string
	ImageURL  string
	Price     string
	SalePrice string
}

// Scraper handles the web scraping configuration and operations
type Scraper struct {
	client    *http.Client
	baseURL   string
	userAgent string
}

// List of common user agents for different browsers and platforms
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2.1 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edge/91.0.864.67",
	"Mozilla/5.0 (Windows NT 6.1; Trident/7.0; AS; en-US) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 OPR/77.0.4054.254",
}

// Rotate User-Agent to make the scraper more stealthy
func (s *Scraper) rotateUserAgent() string {
	// Randomly select a User-Agent from the list
	randomIndex := time.Now().UnixNano() % int64(len(userAgents))
	return userAgents[randomIndex]
}

// NewScraper creates a new scraper instance with sensible defaults
func NewScraper() *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://www.amazon.com",
		userAgent: userAgents[0], // Start with the first User-Agent in the list
	}
}

// ScrapeShirts retrieves shirt listings from Amazon
func (s *Scraper) ScrapeShirts(what string) ([]Product, error) {
	products := make([]Product, 0, 100)

	// Search URL for men's shirts on sale
	searchURL := s.baseURL + "/s?k=" + what + "&s=exact-aware-popularity-rank&qid=1740306793&ref=sr_st_exact-aware-popularity-rank&ds=v1%3AAmHpVBGfprKKIslavORwcpJkRc9A%2FLueerQUpl4iWWU"

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Rotate User-Agent for every request
	s.userAgent = s.rotateUserAgent()

	// Set headers to mimic a real browser
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,es;q=0.8,fr;q=0.7")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Referer", "https://www.amazon.com")
	req.Header.Set("Origin", "https://www.amazon.com")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	fmt.Println(string(data))
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}
	// Find product listings
	doc.Find("div[data-component-type='s-search-result']").Each(func(i int, s *goquery.Selection) {
		if len(products) >= 50 {
			return
		}

		product := Product{}

		// Extract title
		product.Title = strings.TrimSpace(s.Find("h2 span").Text())

		// Extract image URL
		if img := s.Find("img.s-image"); img.Length() > 0 {
			if src, exists := img.Attr("src"); exists {
				product.ImageURL = src
			}
		}

		// Extract price information
		product.Price = strings.TrimSpace(s.Find("span.a-price-whole").First().Text())
		product.SalePrice = strings.TrimSpace(s.Find("span.a-price[data-a-color='secondary'] .a-price-whole").First().Text())

		if product.Title != "" && product.ImageURL != "" {
			products = append(products, product)
		}
	})

	return products, nil
}

func FindTop50(what string) ([]Product, error) {
	scraper := NewScraper()

	products, err := scraper.ScrapeShirts(what)
	if err != nil {
		return nil, err
	}
	return products, nil
}
