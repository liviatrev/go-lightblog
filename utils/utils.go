package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-lightblog/config"
	"go-lightblog/database"
	"go-lightblog/models"

	"github.com/HugoSmits86/nativewebp"
	"github.com/disintegration/imaging"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"github.com/imagekit-developer/imagekit-go/v2/packages/param"
	"google.golang.org/genai"
)

// ============================================================
// 1. CRYPTO / RANDOM GENERATORS
// ============================================================

// GenerateAPIKey creates a random API Key of 32 bytes (64 hex characters)
func GenerateAPIKey() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "api-key-generation-failed"
	}
	return hex.EncodeToString(bytes)
}

// GenerateRandomHex creates a random hex string with a specific byte length
func GenerateRandomHex(byteLen int) string {
	bytes := make([]byte, byteLen)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ============================================================
// 2. SLUG GENERATORS
// ============================================================

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

// GenerateSlug creates a URL slug from input string
// (ROBUST version: lowercase, remove non-alphanumeric, replace with dash)
func GenerateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = slugRegex.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

// GenerateUniqueSlug creates a unique slug in the posts table
// If a slug already exists, add a timestamp suffix
// excludeID parameter is optional: to ignore a specific ID when editing a post
func GenerateUniqueSlug(title string, excludeID ...uint) string {
	slug := GenerateSlug(title)

	var count int64
	query := database.DB.Model(&models.Post{}).Where("slug = ?", slug)

	// If excludeID exists (edit mode), ignore own ID
	if len(excludeID) > 0 {
		query = query.Where("id != ?", excludeID[0])
	}

	query.Count(&count)
	if count > 0 {
		// Add timestamp + random suffix to avoid collision when created in the same second
		slug = slug + "-" + time.Now().Format("150405") + "-" + GenerateRandomHex(2)
	}
	return slug
}

// GenerateSimpleSlug is a simple slug (only replaces spaces with dashes)
// for backward compatibility, use GenerateSlug for new cases
func GenerateSimpleSlug(title string) string {
	return strings.ToLower(strings.ReplaceAll(title, " ", "-"))
}

// ============================================================
// 3. TYPE CONVERSION HELPERS
// ============================================================

// ParseUint converts string to uint (safe without error return)
func ParseUint(s string) uint {
	var n uint
	fmt.Sscanf(s, "%d", &n)
	return n
}

// ParseUintStrict converts string to uint with error handling
func ParseUintStrict(s string) (uint, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

// ============================================================
// 4. SESSION / AUTH HELPERS
// ============================================================

// SessionUser holds user data from session that is frequently accessed
type SessionUser struct {
	Session  *session.Session
	UserID   uint
	UserName string
	UserRole string
}

// GetSessionUser extracts user data from session safely
// (eliminates repeated .(string) casting pattern in all handlers)
func GetSessionUser(c *fiber.Ctx) SessionUser {
	sess, _ := config.SessStore.Get(c)

	var userID uint
	var userName, userRole string

	if uid, ok := sess.Get("user_id").(uint); ok {
		userID = uid
	}
	if name, ok := sess.Get("name").(string); ok {
		userName = name
	}
	if role, ok := sess.Get("role").(string); ok {
		userRole = role
	}

	return SessionUser{
		Session:  sess,
		UserID:   userID,
		UserName: userName,
		UserRole: userRole,
	}
}

// ============================================================
// 5. SITE / CMS SETTING HELPERS
// ============================================================

// GetSiteTitle gets site title from DB with fallback
func GetSiteTitle() string {
	var setting models.Setting
	if err := database.DB.Where("key = ?", "site_title").First(&setting).Error; err != nil {
		return "go-lightblog"
	}
	return setting.Value
}

// GetNavbarData gets Category and Static Page data for navigation menu
func GetNavbarData() fiber.Map {
	var categories []models.Category
	database.DB.Find(&categories)

	var pages []models.Post
	database.DB.Where("is_draft = ? AND type = ?", false, "page").
		Select("title, slug").
		Find(&pages)

	return fiber.Map{
		"NavCategories": categories,
		"NavPages":      pages,
	}
}

// ============================================================
// 6. FILE UPLOAD HELPERS
// ============================================================

// ProcessUpload handles file upload (image) with mode: local or ImageKit CDN
func ProcessUpload(c *fiber.Ctx, formField string) (string, error) {
	fileHeader, err := c.FormFile(formField)
	if err != nil {
		return "", err
	}

	originalFilename := SanitizeFilename(fileHeader.Filename)
	ext := strings.ToLower(filepath.Ext(originalFilename))

	// Only allow raster image types to prevent dangerous file uploads
	// (e.g. HTML/SVG that can be executed when served from static folder)
	if !isAllowedImageExt(ext) {
		return "", fmt.Errorf("file type not allowed")
	}

	nameOnly := strings.TrimSuffix(originalFilename, ext)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	filename := fmt.Sprintf("%s_%s%s", nameOnly, timestamp, ext)

	uploadMode := models.GetSetting(database.DB, "upload_mode", "local")

	if uploadMode == "local" {
		uploadDir := filepath.Join(config.PublicPath, "uploads")
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			return "", err
		}

		savePath := filepath.Join(uploadDir, filename)

		if err := c.SaveFile(fileHeader, savePath); err != nil {
			return "", err
		}

		return "/public/uploads/" + filename, nil
	}

	// ImageKit CDN logic
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	imagekitKey := models.GetSetting(database.DB, "imagekit_private_key", "")
	imagekitFolder := models.GetSetting(database.DB, "imagekit_folder", "/lightblog")
	ik := imagekit.NewClient(
		option.WithPrivateKey(imagekitKey),
	)

	resp, err := ik.Files.Upload(context.TODO(), imagekit.FileUploadParams{
		File:     file,
		FileName: filename,
		Folder:   param.Opt[string]{Value: imagekitFolder},
	})

	if err != nil {
		return "", err
	}

	return resp.URL, nil
}

// SanitizeFilename cleans upload filename from path traversal and dangerous characters.
func SanitizeFilename(name string) string {
	// Normalize Windows path then take only the file component
	base := filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	// Remove remaining control / separator characters
	base = strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', '\x00', '\n', '\r':
			return -1
		}
		return r
	}, base)
	base = strings.TrimSpace(base)
	if base == "" || base == "." || base == ".." {
		return "upload"
	}
	return base
}

// isAllowedImageExt checks if file extension is in the allowed raster image list.
func isAllowedImageExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		return true
	}
	return false
}

// ============================================================
// 7. IMAGE PROCESSING / THUMBNAIL HELPERS
// ============================================================

// ResizeAndCacheThumbnail resizes local image and saves it as cache.
// It always produces both JPG and WebP variants so that HTML templates can
// prioritize WebP with a JPG fallback. The format parameter selects which
// variant path is returned ("jpg" or "webp", default "jpg").
// Used by ImageThumbProxy in handlers/media.go
func ResizeAndCacheThumbnail(src string, w int, format string) (string, error) {
	// Security validation (Prevent Path Traversal)
	if !strings.HasPrefix(src, "/public/uploads/") || strings.Contains(src, "..") {
		return "", fmt.Errorf("access denied")
	}

	// Validate requested output format (jpg or webp), default jpg
	switch strings.ToLower(format) {
	case "", "jpg", "jpeg":
		format = "jpg"
	case "webp":
		format = "webp"
	default:
		return "", fmt.Errorf("unsupported image format")
	}

	// Resolve the local file against the dynamic public folder
	localPath := filepath.Join(config.PublicPath, strings.TrimPrefix(src, "/public/"))

	// Prevent OOM: limit source image size (decompression bomb guard)
	f, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("original image not found")
	}
	cfg, _, err := image.DecodeConfig(f)
	f.Close()
	if err != nil {
		return "", fmt.Errorf("unsupported image format")
	}
	const maxSourceDim = 8000
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > maxSourceDim || cfg.Height > maxSourceDim {
		return "", fmt.Errorf("image too large")
	}

	// Prevent OOM: limit requested thumbnail width (0 < w <= 2000)
	if w <= 0 {
		w = 600
	}
	if w > 2000 {
		w = 2000
	}

	base := strings.TrimSuffix(filepath.Base(localPath), filepath.Ext(localPath))
	thumbDir := filepath.Join(config.PublicPath, "uploads", "thumbs")
	// JPG and WebP variants share the same basename with different extensions
	jpgPath := fmt.Sprintf("%s/w%d_%s.jpg", thumbDir, w, base)
	webpPath := fmt.Sprintf("%s/w%d_%s.webp", thumbDir, w, base)

	thumbPath := jpgPath
	if format == "webp" {
		thumbPath = webpPath
	}

	// 1. Check Cache
	if _, err := os.Stat(thumbPath); err == nil {
		return thumbPath, nil
	}

	// 2. Open original image
	imgFile, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("original image not found")
	}
	defer imgFile.Close()
	img, err := imaging.Decode(imgFile, imaging.AutoOrientation(true))
	if err != nil {
		return "", fmt.Errorf("unsupported image format")
	}

	// 3. Create thumbs folder if not exists
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		return "", fmt.Errorf("failed to save thumbnail")
	}

	// 4. Resize with Lanczos algorithm
	resizedImg := imaging.Resize(img, w, 0, imaging.Lanczos)

	// 5. Ensure both JPG and WebP variants exist
	// (atomic temp-file + rename avoids corruption from concurrent requests)
	if _, err := os.Stat(jpgPath); err != nil {
		if err := saveImageFile(resizedImg, jpgPath, "jpg"); err != nil {
			return "", fmt.Errorf("failed to save thumbnail")
		}
	}
	if _, err := os.Stat(webpPath); err != nil {
		if err := saveImageFile(resizedImg, webpPath, "webp"); err != nil {
			return "", fmt.Errorf("failed to save thumbnail")
		}
	}

	return thumbPath, nil
}

// saveImageFile encodes img into path (jpg or webp) atomically using a
// temp file + rename so a crash never leaves a half-written thumbnail.
func saveImageFile(img image.Image, path, format string) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "thumb-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	var encodeErr error
	switch format {
	case "webp":
		encodeErr = nativewebp.Encode(tmp, img, &nativewebp.Options{})
	default:
		encodeErr = imaging.Encode(tmp, img, imaging.JPEG, imaging.JPEGQuality(70))
	}
	closeErr := tmp.Close()
	if encodeErr != nil || closeErr != nil {
		os.Remove(tmpName)
		return fmt.Errorf("encode failed")
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	// Ensure thumbnail is readable by other processes (e.g. reverse proxy)
	os.Chmod(path, 0644)
	return nil
}

// ============================================================
// 8. SEO / AI HELPERS
// ============================================================

// SEOData output structure from AI SEO Generator
type SEOData struct {
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	TargetKeyword   string `json:"target_keyword"`
}

// SEORequest input structure for generating SEO via API
type SEORequest struct {
	Content string `json:"content"`
}

// ProcessSEOInternal calls Gemini AI to generate SEO metadata from content
func ProcessSEOInternal(content string) (*SEOData, error) {
	apiKey := models.GetSetting(database.DB, "gemini_api_key", "")
	selectedModel := models.GetSetting(database.DB, "gemini_model", "gemini-3.6-flash")

	if apiKey == "" {
		return nil, fmt.Errorf("gemini API Key not configured")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AI client: %v", err)
	}

	prompt := `As an SEO expert, analyze the following article content. 
Generate SEO metadata in the language matching the article.
Return in pure JSON format exactly matching this schema without markdown block:
{"meta_title": "Title max 60 characters", "meta_description": "Description max 160 characters", "target_keyword": "1-4 main keywords"}

Article Content:
` + content

	cfg := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	}

	resp, err := client.Models.GenerateContent(ctx, selectedModel, genai.Text(prompt), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed from AI: %v", err)
	}

	if resp != nil && resp.Text() != "" {
		var seoData SEOData
		err := json.Unmarshal([]byte(resp.Text()), &seoData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON from AI: %v", err)
		}
		return &seoData, nil
	}

	return nil, fmt.Errorf("empty AI response")
}
