package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"strings"

	"go-lightblog/config"
	"go-lightblog/database"
	"go-lightblog/handlers"
	"go-lightblog/models"
	"go-lightblog/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {
	// Parsing command-line flags
	listenAddr := flag.String("l", "127.0.0.1", "Listen address (default 127.0.0.1)")
	listenPort := flag.Int("p", 5800, "Listen port (default 5800)")
	dbPath := flag.String("db", "lightblog.db", "File database")
	flag.Parse()

	// Inisialisasi Database
	database.Connect(*dbPath)

	// Inisialisasi Session Store
	config.InitSession()

	// Inisialisasi Fiber HTML Template Engine
	// Mengarah ke folder "views" dengan ekstensi ".html"
	engine := html.New("./views", ".html")

	// Tambahkan fungsi kustom agar HTML bisa dirender
	engine.AddFunc("unescape", func(s string) template.HTML {
		return template.HTML(s)
	})

	// Tambahkan Fungsi Kustom (Bisa dipanggil di HTML)
	engine.AddFunc("thumb", func(url string, width int) string {
		if url == "" {
			return "/assets/default-cover.jpg" // URL gambar default jika pos tidak punya kover
		}
		
		// Skenario 1: Jika gambar berasal dari ImageKit CDN
		if strings.Contains(url, "imagekit.io") {
			// ImageKit menggunakan parameter query ?tr=w-XXX
			if strings.Contains(url, "?") {
				return fmt.Sprintf("%s&tr=w-%d,q-70,f-webp", url, width)
			}
			return fmt.Sprintf("%s?tr=w-%d,q-70,f-webp", url, width)
		}
		
		// Skenario 2: Jika gambar lokal
		return fmt.Sprintf("/api/thumb?src=%s&w=%d", url, width)
	})

	// Inisialisasi Fiber App
	app := fiber.New(fiber.Config{
		Views:       engine,
		AppName:     "go-lightblog",
		IdleTimeout: 10, // Optimasi ringan untuk manajemen koneksi
	})

	// Daftarkan middleware secara global di sini
	app.Use(middleware.CheckSetup)

	// Melayani file statis dari folder "public"
	app.Static("/public", "./public")

	// Route sementara untuk memastikan server berjalan
	app.Get("/setup", func(c *fiber.Ctx) error {
		return c.Render("setup", fiber.Map{}) // Ubah nil menjadi fiber.Map{}
	})

	// Rute Proses Setup (POST)
	app.Post("/setup/process", handlers.SetupProcess)

	// Rute Login Dinamis
	app.Get("/login-:token", func(c *fiber.Ctx) error {
		// Ambil token dari URL
		urlToken := c.Params("token")
		// Ambil token dari DB (fallback "admin" jika setup belum selesai)
		validToken := models.GetSetting(database.DB, "login_token", "admin")

		// Jika token di URL tidak cocok dengan di DB, lemparkan 404
		if urlToken != validToken {
			return c.Status(fiber.StatusNotFound).SendString("404 Not Found")
		}

		sess, _ := config.SessStore.Get(c)
		if sess.Get("is_logged_in") != nil {
			return c.Redirect("/dashboard")
		}
		
		// Kirim token ke template agar form bisa menggunakan URL aksi yang benar
		return c.Render("login", fiber.Map{
			"LoginToken": urlToken,
		})
	})

	// Proses Login (POST)
	app.Post("/login-:token/process", handlers.ProcessLogin)

	// Rute Halaman Utama (Homepage)
	app.Get("/", handlers.Home)
	app.Get("/post/:slug", handlers.ReadPost)
	app.Get("/page/:slug", handlers.ReadPost) // Untuk Halaman Statis (menggunakan handler yang sama)
	// Rute Publik (Tambahkan di bawah rute GET lainnya)
	app.Get("/api/thumb", handlers.ImageThumbProxy)
	// Rute Pencarian Publik
	app.Get("/search", handlers.SearchPosts)

	// Rute Arsip (Kategori, Tag, Author)
	app.Get("/category/:slug", handlers.CategoryPosts)
	app.Get("/tag/:slug", handlers.TagPosts)
	app.Get("/author/:id", handlers.AuthorPosts)

	// ==========================================
	// API MANAJEMEN CMS (KHUSUS ADMIN / EDITOR)
	// ==========================================
	
	// Semua rute di bawah /api/v1/admin wajib login
	adminAPI := app.Group("/api/v1/admin", middleware.RequireAuth)

	// 1. CRUD Pengguna (Hanya Admin)
	users := adminAPI.Group("/users", middleware.RequireRole("admin"))
	users.Get("/", handlers.ApiGetUsers)
	users.Post("/", handlers.ApiCreateUser)
	users.Put("/:id", handlers.ApiUpdateUser)
	users.Delete("/:id", handlers.ApiDeleteUser)

	// 2. Read/Write Settings (Hanya Admin)
	settings := adminAPI.Group("/settings", middleware.RequireRole("admin"))
	settings.Get("/", handlers.ApiGetSettings)
	settings.Put("/", handlers.ApiUpdateSettings) // Gunakan PUT/PATCH karena setting biasanya entitas tunggal

	// 3. CRUD Kategori & Tags (Admin & Editor)
	// Editor bisa membuat tag/kategori baru untuk artikel mereka
	categories := adminAPI.Group("/categories", middleware.RequireRole("admin", "editor"))
	categories.Get("/", handlers.ApiGetCategories)
	categories.Post("/", handlers.ApiCreateCategory)
	categories.Put("/:id", handlers.ApiUpdateCategory)
	categories.Delete("/:id", handlers.ApiDeleteCategory)

	tags := adminAPI.Group("/tags", middleware.RequireRole("admin", "editor"))
	tags.Get("/", handlers.ApiGetTags)
	tags.Post("/", handlers.ApiCreateTag)
	tags.Put("/:id", handlers.ApiUpdateTag)
	tags.Delete("/:id", handlers.ApiDeleteTag)

	// // 4. CRUD Posts + Cover Upload + AI SEO (Admin & Editor)
	posts := adminAPI.Group("/posts", middleware.RequireRole("admin", "editor"))
	posts.Get("/", handlers.ApiGetPosts)
	posts.Post("/", handlers.ApiCreatePost) // Di sini akan ditangani file upload (multipart/form-data) dan AI trigger
	posts.Put("/:id", handlers.ApiUpdatePost)
	posts.Delete("/:id", handlers.ApiDeletePost)

	adminGroup := app.Group("/", middleware.RequireLogin)

	adminGroup.Get("/dashboard", handlers.DashboardView)

	// --- ROUTE POSTINGAN ---
	adminGroup.Get("/posts", handlers.ListPosts)             // Tampilkan daftar
	adminGroup.Get("/posts/create", handlers.CreatePostView) // Tampilkan form
	adminGroup.Post("/posts/create", handlers.ProcessCreatePost) // Proses simpan
	adminGroup.Post("/seo/generate", handlers.GenerateSEO)

	// Rute Edit & Delete baru
	adminGroup.Get("/posts/edit/:id", handlers.EditPostView)
	adminGroup.Post("/posts/edit/:id", handlers.ProcessEditPost)
	adminGroup.Get("/posts/delete/:id", handlers.DeletePost)

	// Manajemen Kategori
	adminGroup.Get("/categories", handlers.CategoryList)
	adminGroup.Post("/categories", handlers.CategoryCreate)
	adminGroup.Post("/categories/delete/:id", handlers.CategoryDelete)

	// Manajemen Tag
	adminGroup.Get("/tags", handlers.TagList)
	adminGroup.Post("/tags", handlers.TagCreate)
	adminGroup.Post("/tags/delete/:id", handlers.TagDelete)

	// API Internal untuk Upload
	adminGroup.Post("/api/upload", handlers.UploadImage)

	// Proses Logout (GET/POST - untuk sekarang GET lebih praktis diujicoba)
    adminGroup.Post("/logout", handlers.ProcessLogout)

	// === TAMBAHKAN GRUP SUPER ADMIN ===
	// Rute Khusus Admin
	superAdminGroup := app.Group("", middleware.RequireLogin, middleware.RequireAdmin)

	// --- ROUTE PENGATURAN ---
	superAdminGroup.Get("/settings", handlers.SettingsView)
	superAdminGroup.Post("/settings/update", handlers.ProcessUpdateSettings)
	superAdminGroup.Post("/settings/password", handlers.ProcessUpdatePassword)
	superAdminGroup.Post("/settings/integrations", handlers.ProcessUpdateIntegrations)

	// Rute Manajemen Pengguna (Hanya Admin)
	superAdminGroup.Get("/users", handlers.ListUsers)
	superAdminGroup.Get("/users/create", handlers.CreateUserView)
	superAdminGroup.Post("/users/create", handlers.CreateUserProcess)
	superAdminGroup.Post("/users/delete/:id", handlers.DeleteUserProcess)


	// Menyiapkan alamat dan menjalankan server
	addr := fmt.Sprintf("%s:%d", *listenAddr, *listenPort)
	log.Printf("Starting go-lightblog on http://%s", addr)
	
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}