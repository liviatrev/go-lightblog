package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

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

	// Initialize Database
	database.Connect(*dbPath)

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
	engine.AddFunc("thumb", func(url string, width int) string {
		if url == "" {
			return "/public/assets/default-cover.jpg" // Default image URL if post has no cover
		}
		
		// Scenario 1: If image comes from ImageKit CDN
		if strings.Contains(url, "imagekit.io") {
			// ImageKit uses query parameter ?tr=w-XXX
			if strings.Contains(url, "?") {
				return fmt.Sprintf("%s&tr=w-%d,q-70,f-webp", url, width)
			}
			return fmt.Sprintf("%s?tr=w-%d,q-70,f-webp", url, width)
		}
		
		// Scenario 2: If image is local
		return fmt.Sprintf("/api/thumb?src=%s&w=%d", url, width)
	})

	// Initialize Fiber App
	app := fiber.New(fiber.Config{
		Views:       engine,
		AppName:     "go-lightblog",
		IdleTimeout: 10 * time.Second, // Light optimization for connection management
	})

	// Register middleware globally here
	app.Use(middleware.CheckSetup)

	// Serve static files from "public" folder
	app.Static("/public", "./public")

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


	// Set up address and run server
	addr := fmt.Sprintf("%s:%d", *listenAddr, *listenPort)
	log.Printf("Starting go-lightblog on http://%s", addr)
	
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
