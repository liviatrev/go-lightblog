package handlers

import (
	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// ========================================================
// 1. GET ALL USERS (Read)
// ========================================================
func ApiGetUsers(c *fiber.Ctx) error {
	var users []models.User

	// Get all user data.
	// Make sure in models.User, the password column already uses json:"-" tag
	if err := database.DB.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch user data",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    users,
	})
}

// ========================================================
// 2. CREATE USER (With Password Hashing)
// ========================================================
type CreateUserRequest struct {
	Username string `json:"username" form:"username"`
	Name     string `json:"name" form:"name"`
	Password string `json:"password" form:"password"`
	Role     string `json:"role" form:"role"`
	APIKey   string `json:"api_key" form:"api_key"`
}

func ApiCreateUser(c *fiber.Ctx) error {
	req := new(CreateUserRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request format",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Username and Password are required",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to process password",
		})
	}

	newUser := models.User{
		Username: req.Username,
		Name:     req.Name,
		Password: string(hashedPassword),
		Role:     req.Role,
		APIKey:   utils.GenerateAPIKey(),
	}

	if newUser.Role == "" {
		newUser.Role = "editor"
	}

	// Save to Database
	if err := database.DB.Create(&newUser).Error; err != nil {
		// Assume error due to duplication (Username already registered)
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create user. Username may already be in use.",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User added successfully!",
		"data":    newUser, // Password automatically hidden thanks to json:"-" in model struct
	})
}

// ========================================================
// 3. UPDATE USER (Partial)
// ========================================================
func ApiUpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User

	// 1. Make sure user exists
	if err := database.DB.First(&user, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
		})
	}

	// 2. Parse request JSON into a flexible Map form
	var input map[string]interface{}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request format. Make sure to send JSON format.",
		})
	}

	// 3. Prepare data container for GORM (columns mapped to database)
	updateData := make(map[string]interface{})

	if val, ok := input["username"]; ok {
		updateData["username"] = val
	}
	if val, ok := input["name"]; ok {
		updateData["name"] = val
	}
	if val, ok := input["role"]; ok {
		updateData["role"] = val
	}

	// 4. Special Handling: Hash Password if user wants to change it
	if val, ok := input["password"]; ok {
		passwordStr, isString := val.(string)

		// Only process if password is a string and not empty
		if isString && passwordStr != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordStr), bcrypt.DefaultCost)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Failed to process new password",
				})
			}
			updateData["password"] = string(hashedPassword)
		}
	}

	// 5. Execute Update to Database ONLY if there is data in container
	if len(updateData) > 0 {
		if err := database.DB.Model(&user).Updates(updateData).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to update user data. Username may already be in use.",
			})
		}
	}

	// 6. Reload latest data (so response doesn't return stale data)
	database.DB.First(&user, id)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User data updated successfully!",
		"data":    user,
	})
}

// ========================================================
// 4. DELETE USER
// ========================================================
func ApiDeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User

	if err := database.DB.First(&user, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
		})
	}

	// Delete user
	// Note: If you add gorm.DeletedAt in models.User,
	// this automatically becomes Soft Delete like on Posts.
	if err := database.DB.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete user",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User deleted successfully",
	})
}
