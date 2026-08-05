package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-lightblog/config"
	"go-lightblog/database"
	"go-lightblog/models"

	"github.com/disintegration/imaging"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"github.com/imagekit-developer/imagekit-go/v2/packages/param"
	"google.golang.org/genai"
	"encoding/json"
)

// ============================================================
// 1. CRYPTO / RANDOM GENERATORS
// ============================================================

// GenerateAPIKey membuat API Key acak sepanjang 32 byte (64 karakter hex)
func GenerateAPIKey() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "api-key-generation-failed"
	}
	return hex.EncodeToString(bytes)
}

// GenerateRandomHex membuat string hex acak dengan panjang bytes tertentu
func GenerateRandomHex(byteLen int) string {
	bytes := make([]byte, byteLen)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ============================================================
// 2. SLUG GENERATORS
// ============================================================

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

// GenerateSlug membuat slug URL dari string input
// (versi ROBUST: lowercase, hapus non-alphanumeric, ganti dengan strip)
func GenerateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = slugRegex.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

// GenerateUniqueSlug membuat slug yang unik di tabel posts
// Jika sudah ada slug yang sama, tambahkan suffix timestamp
// Parameter excludeID opsional: untuk mengabaikan ID tertentu saat edit post
func GenerateUniqueSlug(title string, excludeID ...uint) string {
	slug := GenerateSlug(title)

	var count int64
	query := database.DB.Model(&models.Post{}).Where("slug = ?", slug)

	// Jika ada excludeID (mode edit), abaikan ID sendiri
	if len(excludeID) > 0 {
		query = query.Where("id != ?", excludeID[0])
	}

	query.Count(&count)
	if count > 0 {
		slug = slug + "-" + time.Now().Format("150405")
	}
	return slug
}

// GenerateSimpleSlug adalah slug sederhana (hanya replace spasi dengan strip)
// untuk kompatibilitas lama, sebaiknya gunakan GenerateSlug untuk kasus baru
func GenerateSimpleSlug(title string) string {
	return strings.ToLower(strings.ReplaceAll(title, " ", "-"))
}

// ============================================================
// 3. TYPE CONVERSION HELPERS
// ============================================================

// ParseUint konversi string ke uint (aman tanpa error return)
func ParseUint(s string) uint {
	var n uint
	fmt.Sscanf(s, "%d", &n)
	return n
}

// ParseUintStrict konversi string ke uint dengan error handling
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

// SessionUser menampung data user dari session yang sering diakses
type SessionUser struct {
	Session  *session.Session
	UserID   uint
	UserName string
	UserRole string
}

// GetSessionUser mengekstrak data user dari session secara aman
// (menghilangkan pola .(string) casting berulang di semua handler)
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

// GetSiteTitle mengambil judul situs dari DB dengan fallback
func GetSiteTitle() string {
	var setting models.Setting
	if err := database.DB.Where("key = ?", "site_title").First(&setting).Error; err != nil {
		return "go-lightblog"
	}
	return setting.Value
}

// GetNavbarData mengambil data Kategori dan Halaman Statis untuk menu navigasi
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

// ProcessUpload menangani upload file (gambar) dengan mode: local atau ImageKit CDN
func ProcessUpload(c *fiber.Ctx, formField string) (string, error) {
	fileHeader, err := c.FormFile(formField)
	if err != nil {
		return "", err
	}

	originalFilename := fileHeader.Filename
	ext := filepath.Ext(originalFilename)
	nameOnly := strings.TrimSuffix(originalFilename, ext)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	filename := fmt.Sprintf("%s_%s%s", nameOnly, timestamp, ext)

	uploadMode := models.GetSetting(database.DB, "upload_mode", "local")

	if uploadMode == "local" {
		uploadDir := "./public/uploads"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			return "", err
		}

		savePath := filepath.Join(uploadDir, filename)

		if err := c.SaveFile(fileHeader, savePath); err != nil {
			return "", err
		}

		return "/public/uploads/" + filename, nil
	}

	// Logika ImageKit CDN
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

// ============================================================
// 7. IMAGE PROCESSING / THUMBNAIL HELPERS
// ============================================================

// ResizeAndCacheThumbnail meresize gambar lokal dan menyimpannya sebagai cache
// Dipakai oleh ImageThumbProxy di handlers/media.go
func ResizeAndCacheThumbnail(src string, w int) (string, error) {
	// Validasi keamanan (Mencegah Path Traversal)
	if !strings.HasPrefix(src, "/public/uploads/") || strings.Contains(src, "..") {
		return "", fmt.Errorf("akses ditolak")
	}

	localPath := "." + src
	fileName := filepath.Base(localPath)
	thumbDir := "./public/uploads/thumbs"
	thumbPath := fmt.Sprintf("%s/w%d_%s", thumbDir, w, fileName)

	// 1. Cek Cache
	if _, err := os.Stat(thumbPath); err == nil {
		return thumbPath, nil
	}

	// 2. Buka gambar asli
	img, err := imaging.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("gambar asli tidak ditemukan")
	}

	// 3. Buat folder thumbs jika belum ada
	os.MkdirAll(thumbDir, 0755)

	// 4. Resize dengan algoritma Lanczos
	resizedImg := imaging.Resize(img, w, 0, imaging.Lanczos)

	// 5. Simpan thumbnail dengan kompresi JPEG 70%
	err = imaging.Save(resizedImg, thumbPath, imaging.JPEGQuality(70))
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan thumbnail")
	}

	return thumbPath, nil
}

// ============================================================
// 8. SEO / AI HELPERS
// ============================================================

// SEOData struktur output dari AI SEO Generator
type SEOData struct {
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	TargetKeyword   string `json:"target_keyword"`
}

// SEORequest struktur input untuk generate SEO via API
type SEORequest struct {
	Content string `json:"content"`
}

// ProcessSEOInternal memanggil Gemini AI untuk generate metadata SEO dari konten
func ProcessSEOInternal(content string) (*SEOData, error) {
	apiKey := models.GetSetting(database.DB, "gemini_api_key", "")
	selectedModel := models.GetSetting(database.DB, "gemini_model", "gemini-3.6-flash")

	if apiKey == "" {
		return nil, fmt.Errorf("gemini API Key belum diatur")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("gagal inisialisasi AI client: %v", err)
	}

	prompt := `Sebagai pakar SEO, analisis konten artikel berikut. 
Hasilkan metadata SEO dalam bahasa yang sesuai dengan artikel.
Kembalikan dalam format JSON murni persis dengan skema ini tanpa markdown block:
{"meta_title": "Judul maksimal 60 karakter", "meta_description": "Deskripsi maksimal 160 karakter", "target_keyword": "1-4 kata kunci utama"}

Konten Artikel:
` + content

	cfg := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	}

	resp, err := client.Models.GenerateContent(ctx, selectedModel, genai.Text(prompt), cfg)
	if err != nil {
		return nil, fmt.Errorf("gagal dari AI: %v", err)
	}

	if resp != nil && resp.Text() != "" {
		var seoData SEOData
		err := json.Unmarshal([]byte(resp.Text()), &seoData)
		if err != nil {
			return nil, fmt.Errorf("gagal mem-parsing JSON dari AI: %v", err)
		}
		return &seoData, nil
	}

	return nil, fmt.Errorf("respon AI kosong")
}
