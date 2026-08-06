package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"go-lightblog/config"
	"go-lightblog/database"
	"go-lightblog/handlers"
	"go-lightblog/models"
	"go-lightblog/middleware"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

// ThumbURLs holds both variants produced by the "thumb" template function,
// letting templates prioritize WebP with a JPG fallback.
type ThumbURLs struct {
	WebP string
	JPG  string
}

func main() {
	// Parsing command-line flags
	listenAddr := flag.String("l", "", "Listen TCP address (e.g., 127.0.0.1)")
	listenPort := flag.Int("p", 0, "Listen TCP port (default 5800)")
	socketPath := flag.String("sock", "/tmp/lightblog.sock", "Unix socket path (default if no TCP flags are provided)")
	dbPath := flag.String("db", "lightblog.db", "File database")
	publicPath := flag.String("a", "./public", "Public folder")
	flag.Parse()

	// Initialize Database
	database.Connect(*dbPath)

	// Pass the public folder flag to helpers (uploads / thumbnails storage)
	config.PublicPath = *publicPath

	// Initialize Session Store
	config.InitSession()

	// Initialize Fiber HTML Template Engine
	// Points to "views" folder with ".html" extension
	engine := html.New("./views", ".html")

	// Add custom function so HTML can be rendered
	engine.AddFunc("unescape", func(s string) template.HTML {
		return template.HTML(s)
	})

	// Add Custom Function (Can be called in HTML)
	engine.AddFunc("thumb", func(url string, width int) ThumbURLs {
		if url == "" {
			// Default image URL if post has no cover
			return ThumbURLs{
				WebP: "/public/assets/default-cover.jpg",
				JPG:  "/public/assets/default-cover.jpg",
			}
		}

		// Scenario 1: If image comes from ImageKit CDN
		if strings.Contains(url, "imagekit.io") {
			// ImageKit uses query parameter ?tr=w-XXX
			sep := "?"
			if strings.Contains(url, "?") {
				sep = "&"
			}
			return ThumbURLs{
				WebP: fmt.Sprintf("%s%str=w-%d,q-70,f-webp", url, sep, width),
				JPG:  fmt.Sprintf("%s%str=w-%d,q-70,f-jpg", url, sep, width),
			}
		}

		// Scenario 2: If image is local
		return ThumbURLs{
			WebP: fmt.Sprintf("/api/thumb?src=%s&w=%d&f=webp", url, width),
			JPG:  fmt.Sprintf("/api/thumb?src=%s&w=%d&f=jpg", url, width),
		}
	})

	// Add Cache Buster Functions for static files
	engine.AddFunc("cacheBuster", utils.GetCacheBuster)
	engine.AddFunc("cacheBusterURL", utils.CacheBusterURL)

	// Initialize Fiber App
	app := fiber.New(fiber.Config{
		Views:       engine,
		AppName:     "go-lightblog",
		IdleTimeout: 10 * time.Second, // Light optimization for connection management
	})

	// Register middleware globally here
	app.Use(middleware.CheckSetup)

	// Serve static files from "public" folder
	app.Static("/public", *publicPath)

	// Temporary route to ensure server is running
	app.Get("/setup", func(c *fiber.Ctx) error {
		return c.Render("setup", fiber.Map{}) // Change nil to fiber.Map{}
	})

	// Setup Process Route (POST)
	app.Post("/setup/process", handlers.SetupProcess)

	// Dynamic Login Route
	app.Get("/login-:token", func(c *fiber.Ctx) error {
		// Get token from URL
		urlToken := c.Params("token")
		// Get token from DB (fallback "admin" if setup not completed)
		validToken := models.GetSetting(database.DB, "login_token", "admin")

		// If URL token doesn't match DB, throw 404
		if urlToken != validToken {
			return c.Status(fiber.StatusNotFound).SendString("404 Not Found")
		}

		sess, _ := config.SessStore.Get(c)
		if sess.Get("is_logged_in") != nil {
			return c.Redirect("/dashboard")
		}
		
		// Send token to template so form can use correct action URL
		return c.Render("login", fiber.Map{
			"LoginToken": urlToken,
		})
	})

	// Login Process (POST)
	app.Post("/login-:token/process", handlers.ProcessLogin)

	// Homepage Routes
	app.Get("/", handlers.Home)
	app.Get("/post/:slug", handlers.ReadPost)
	app.Get("/page/:slug", handlers.ReadPost) // For Static Pages (using same handler)
	// Public Routes (Add below other GET routes)
	app.Get("/api/thumb", handlers.ImageThumbProxy)
	// Public Search Route
	app.Get("/search", handlers.SearchPosts)

	// Archive Routes (Category, Tag, Author)
	app.Get("/category/:slug", handlers.CategoryPosts)
	app.Get("/tag/:slug", handlers.TagPosts)
	app.Get("/author/:id", handlers.AuthorPosts)

	// ==========================================
	// CMS MANAGEMENT API (ADMIN / EDITOR ONLY)
	// ==========================================
	
	// All routes under /api/v1/admin require login
	adminAPI := app.Group("/api/v1/admin", middleware.RequireAuth)

	// 1. User CRUD (Admin only)
	users := adminAPI.Group("/users", middleware.RequireRole("admin"))
	users.Get("/", handlers.ApiGetUsers)
	users.Post("/", handlers.ApiCreateUser)
	users.Put("/:id", handlers.ApiUpdateUser)
	users.Delete("/:id", handlers.ApiDeleteUser)

	// 2. Read/Write Settings (Admin only)
	settings := adminAPI.Group("/settings", middleware.RequireRole("admin"))
	settings.Get("/", handlers.ApiGetSettings)
	settings.Put("/", handlers.ApiUpdateSettings) // Use PUT/PATCH because setting is usually a single entity

	// 3. Category & Tag CRUD (Admin & Editor)
	// Editor can create new tags/categories for their articles
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

	// // 4. Post CRUD + Cover Upload + AI SEO (Admin & Editor)
	posts := adminAPI.Group("/posts", middleware.RequireRole("admin", "editor"))
	posts.Get("/", handlers.ApiGetPosts)
	posts.Post("/", handlers.ApiCreatePost) // File upload (multipart/form-data) and AI trigger handled here
	posts.Put("/:id", handlers.ApiUpdatePost)
	posts.Delete("/:id", handlers.ApiDeletePost)

	adminGroup := app.Group("/", middleware.RequireLogin)

	adminGroup.Get("/dashboard", handlers.DashboardView)

	// --- POST ROUTES ---
	adminGroup.Get("/posts", handlers.ListPosts)             // Show list
	adminGroup.Get("/posts/create", handlers.CreatePostView) // Show form
	adminGroup.Post("/posts/create", handlers.ProcessCreatePost) // Process save
	adminGroup.Post("/seo/generate", handlers.GenerateSEO)

	// New Edit & Delete routes
	adminGroup.Get("/posts/edit/:id", handlers.EditPostView)
	adminGroup.Post("/posts/edit/:id", handlers.ProcessEditPost)
	adminGroup.Get("/posts/delete/:id", handlers.DeletePost)

	// Category Management
	adminGroup.Get("/categories", handlers.CategoryList)
	adminGroup.Post("/categories", handlers.CategoryCreate)
	adminGroup.Post("/categories/delete/:id", handlers.CategoryDelete)

	// Tag Management
	adminGroup.Get("/tags", handlers.TagList)
	adminGroup.Post("/tags", handlers.TagCreate)
	adminGroup.Post("/tags/delete/:id", handlers.TagDelete)

	// Internal API for Upload
	adminGroup.Post("/api/upload", handlers.UploadImage)

	// Logout Process (GET/POST - for now GET is more practical for testing)
    adminGroup.Post("/logout", handlers.ProcessLogout)

	// === ADD SUPER ADMIN GROUP ===
	// Admin-only routes
	superAdminGroup := app.Group("", middleware.RequireLogin, middleware.RequireAdmin)

	// --- SETTINGS ROUTES ---
	superAdminGroup.Get("/settings", handlers.SettingsView)
	superAdminGroup.Post("/settings/update", handlers.ProcessUpdateSettings)
	superAdminGroup.Post("/settings/password", handlers.ProcessUpdatePassword)
	superAdminGroup.Post("/settings/integrations", handlers.ProcessUpdateIntegrations)

	// User Management Routes (Admin only)
	superAdminGroup.Get("/users", handlers.ListUsers)
	superAdminGroup.Get("/users/create", handlers.CreateUserView)
	superAdminGroup.Post("/users/create", handlers.CreateUserProcess)
	superAdminGroup.Post("/users/delete/:id", handlers.DeleteUserProcess)


	// ==========================================
	// SERVER LISTENER LOGIC
	// ==========================================

	// If there is flag -l (listenAddr) or -p (listenPort), use TCP mode
	if *listenAddr != "" || *listenPort != 0 {
		host := "127.0.0.1" // Fallback for host
		if *listenAddr != "" {
			host = *listenAddr
		}
		
		port := 5800 // Fallback for port
		if *listenPort != 0 {
			port = *listenPort
		}
		
		addr := fmt.Sprintf("%s:%d", host, port)
		log.Printf("Starting go-lightblog on TCP http://%s", addr)
		
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Error starting TCP server: %v", err)
		}
	} else {
		// If no TCP params, use Unix Socket
		log.Printf("Starting go-lightblog on Unix Socket: %s", *socketPath)

		// 1. Clean old socket file
		if _, err := os.Stat(*socketPath); err == nil {
			if err := os.Remove(*socketPath); err != nil {
				log.Fatalf("Failed to remove existing socket file: %v", err)
			}
		}

		// 2. Create new listener
		ln, err := net.Listen("unix", *socketPath)
		if err != nil {
			log.Fatalf("Failed to listen on unix socket: %v", err)
		}

		// 3. Set permission for proxy server (www-data/caddy)
		if err := os.Chmod(*socketPath, 0666); err != nil {
			log.Printf("Warning: Failed to set socket permissions: %v", err)
		}

		// 4. run through custom listener
		if err := app.Listener(ln); err != nil {
			log.Fatalf("Error starting Unix Socket server: %v", err)
		}
	}
}
