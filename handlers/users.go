package handlers

import (
	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func ListUsers(c *fiber.Ctx) error {
	var users []models.User
	database.DB.Order("role asc").Find(&users)

	su := utils.GetSessionUser(c)

	return c.Render("dashboard/users_list", fiber.Map{
		"Title":       "Manajemen Pengguna",
		"HeaderTitle": "Daftar Pengguna",
		"ActiveMenu":  "users",
		"Users":       users,
		"UserName":    su.UserName,
		"UserRole":    su.UserRole,
	}, "layouts/main")
}

func CreateUserView(c *fiber.Ctx) error {
	su := utils.GetSessionUser(c)

	return c.Render("dashboard/user_form", fiber.Map{
		"Title":       "Tambah Pengguna",
		"HeaderTitle": "Tambah Editor Baru",
		"ActiveMenu":  "users",
		"UserName":    su.UserName,
		"UserRole":    su.UserRole,
	}, "layouts/main")
}

func CreateUserProcess(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	name := c.FormValue("name")
	role := c.FormValue("role")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).SendString("Kesalahan Server Internal (Bcrypt)")
	}

	user := models.User{
		Username: username,
		Password: string(hashedPassword),
		Name:     name,
		Role:     role,
		APIKey:   utils.GenerateAPIKey(),
	}

	database.DB.Create(&user)
	return c.Redirect("/users")
}

func DeleteUserProcess(c *fiber.Ctx) error {
	id := c.Params("id")
	su := utils.GetSessionUser(c)

	var userToDelete models.User
	database.DB.First(&userToDelete, id)

	if userToDelete.ID == su.UserID {
		return c.Redirect("/users")
	}

	database.DB.Delete(&models.User{}, id)
	return c.Redirect("/users")
}