package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// cacheBusterCache stores computed file hashes to avoid recomputing on every request
var cacheBusterCache = make(map[string]string)
var cacheBusterMutex sync.RWMutex

// GetCacheBuster returns a cache buster string for a static file
// It computes an MD5 hash of the file content and returns the first 8 characters
// If the file doesn't exist, it returns an empty string
func GetCacheBuster(filePath string) string {
	// Check cache first
	cacheBusterMutex.RLock()
	if cached, ok := cacheBusterCache[filePath]; ok {
		cacheBusterMutex.RUnlock()
		return cached
	}
	cacheBusterMutex.RUnlock()

	// Convert public path to actual file path
	// e.g., "/public/css/public.min.css" -> "./public/css/public.min.css"
	actualPath := "." + filePath

	// Security: prevent path traversal
	if strings.Contains(filePath, "..") {
		return ""
	}

	// Open file
	file, err := os.Open(actualPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Compute MD5 hash
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	// Take first 8 characters of hex hash
	buster := fmt.Sprintf("%x", hash.Sum(nil))[:8]

	// Store in cache
	cacheBusterMutex.Lock()
	cacheBusterCache[filePath] = buster
	cacheBusterMutex.Unlock()

	return buster
}

// CacheBusterURL returns a URL with cache buster query parameter
// e.g., "/public/css/public.min.css" -> "/public/css/public.min.css?v=abc12345"
func CacheBusterURL(filePath string) string {
	buster := GetCacheBuster(filePath)
	if buster == "" {
		return filePath
	}
	return filePath + "?v=" + buster
}

// ClearCacheBusterCache clears the cache buster cache (useful for development/testing)
func ClearCacheBusterCache() {
	cacheBusterMutex.Lock()
	defer cacheBusterMutex.Unlock()
	cacheBusterCache = make(map[string]string)
}