package Crawler

import (
	"errors"
	"io"
	"net/http"

	"gocv.io/x/gocv"
)

func downloadImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func OrbCheckURL(img1_url, img2_url string) (bool, error) {
	img1, err := downloadImage(img1_url)
	if err != nil {
		return false, err
	}
	img2, err := downloadImage(img2_url)
	if err != nil {
		return false, err
	}
	return orb(img1, img2)

}

func orb(img1_file, img2_file []byte) (bool, error) {
	// Load images
	img1, err1 := gocv.IMDecode(img1_file, gocv.IMReadGrayScale)
	img2, err2 := gocv.IMDecode(img2_file, gocv.IMReadGrayScale)
	if err1 != nil || err2 != nil {
		return false, errors.New("failed Loading Imange")
	}
	defer img1.Close()
	defer img2.Close()

	// Create ORB detector
	orb := gocv.NewORB()
	defer orb.Close()

	// Detect keypoints and compute descriptors
	_, des1 := orb.DetectAndCompute(img1, gocv.NewMat())
	_, des2 := orb.DetectAndCompute(img2, gocv.NewMat())

	// Check if descriptors are valid
	if des1.Empty() || des2.Empty() {
		return false, errors.New("error computing descriptors")
	}

	// Create BFMatcher
	matcher := gocv.NewBFMatcher()
	defer matcher.Close()

	// Match descriptors
	matches := matcher.KnnMatch(des1, des2, 2)

	// Apply Lowe’s ratio test
	goodMatches := 0
	for _, m := range matches {
		if len(m) == 2 && m[0].Distance < 0.75*m[1].Distance {
			goodMatches++
		}
	}
	return goodMatches > 10, nil
}

// func test_obs_check() {
// 	good, err := orb("shirt.jpg", "proxy-image.jpg")
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	if good { // Adjust threshold as needed
// 		fmt.Println("Images are similar")
// 	} else {
// 		fmt.Println("Images are not similar")
// 	}
// }
