// utils/indexnow.go
package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-lightblog/database"
	"go-lightblog/models"
)

// ============================================================
// INDEXNOW SUBMISSION HELPERS
// ============================================================

const indexNowAPI = "https://api.indexnow.org/indexnow"

// GenerateIndexNowKey creates a random alphanumeric key of 80-128 characters
// (a-zA-Z0-9) as required by the IndexNow protocol.
func GenerateIndexNowKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := 80 + int(randInt(49)) // 80..128
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[randInt(int64(len(charset)))]
	}
	return string(b)
}

// randInt returns a cryptographically secure random integer in [0, max).
func randInt(max int64) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0
	}
	return n.Int64()
}

// IsIndexNowEnabled checks if the IndexNow submission feature is enabled.
func IsIndexNowEnabled() bool {
	return models.GetSetting(database.DB, "indexnow", "no") == "yes"
}

// GetIndexNowKey returns the configured IndexNow key.
func GetIndexNowKey() string {
	return models.GetSetting(database.DB, "indexnow_key", "")
}

// GetIndexNowHost returns the host (domain without scheme) used for the
// IndexNow JSON submission. It derives the host from the site_url setting.
// Returns empty string if site_url is not configured.
func GetIndexNowHost() string {
	siteURL := GetSiteURL()
	if siteURL == "" {
		return ""
	}

	// Strip scheme (https://, http://) and trailing slashes
	host := strings.TrimPrefix(siteURL, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.Trim(host, "/")

	// Remove any path component (keep only host:port)
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}

	return host
}

// indexNowInitialPayload is the JSON body for the initial IndexNow submission.
type indexNowInitialPayload struct {
	Host    string   `json:"host"`
	Key     string   `json:"key"`
	URLList []string `json:"urlList"`
}

// SubmitIndexNowInitial submits the homepage URL to IndexNow the first time
// the feature is enabled. It only runs once (guarded by indexnow_submitted).
// Non-blocking: runs in a goroutine so it never blocks the response path.
func SubmitIndexNowInitial() {
	if !IsIndexNowEnabled() {
		return
	}

	// Only submit once
	if models.GetSetting(database.DB, "indexnow_submitted", "no") == "yes" {
		return
	}

	host := GetIndexNowHost()
	key := GetIndexNowKey()
	if host == "" || key == "" {
		return
	}

	payload := indexNowInitialPayload{
		Host:    "https://" + host,
		Key:     key,
		URLList: []string{"https://" + host},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("IndexNow initial submit marshal error: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, indexNowAPI, bytes.NewReader(body))
	if err != nil {
		log.Printf("IndexNow initial submit request error: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("IndexNow initial submit error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("IndexNow initial submit returned status %d", resp.StatusCode)
		return
	}

	// Mark as submitted so we don't resubmit on every restart
	database.DB.Save(&models.Setting{
		Key:   "indexnow_submitted",
		Value: "yes",
	})
	log.Printf("IndexNow initial submission completed for host %s", host)
}

// SubmitIndexNowURL submits a single post/page URL to IndexNow.
// The URL is sent as a simple GET-style POST request:
//   https://api.indexnow.org/indexnow?url={url_encoded}&key={key}
//
// Non-blocking: runs in a goroutine so it never blocks the response path.
func SubmitIndexNowURL(post models.Post) {
	if !IsIndexNowEnabled() {
		return
	}

	key := GetIndexNowKey()
	siteURL := GetSiteURL()
	if key == "" || siteURL == "" {
		return
	}

	// Normalize base URL (strip trailing slash)
	for len(siteURL) > 1 && siteURL[len(siteURL)-1] == '/' {
		siteURL = siteURL[:len(siteURL)-1]
	}

	// Build the post/page URL
	var postURL string
	if post.Type == "page" {
		postURL = siteURL + "/page/" + post.Slug
	} else {
		postURL = siteURL + "/post/" + post.Slug
	}

	// Build the IndexNow endpoint with url-encoded URL and key
	endpoint := fmt.Sprintf("%s?url=%s&key=%s", indexNowAPI, url.QueryEscape(postURL), url.QueryEscape(key))

	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		log.Printf("IndexNow URL submit request error: %v", err)
		return
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("IndexNow URL submit error for %s: %v", postURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("IndexNow URL submit returned status %d for %s", resp.StatusCode, postURL)
		return
	}

	log.Printf("IndexNow URL submitted: %s", postURL)
}