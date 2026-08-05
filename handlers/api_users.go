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

	// Mengambil semua data pengguna.
	// Pastikan pada models.User, kolom password sudah memakai tag json:"-"
	if err := database.DB.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data pengguna",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    users,
	})
}

// ========================================================
// 2. CREATE USER (Dengan Hashing Password)
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
			"message": "Format request tidak valid",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Username dan Password wajib diisi",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memproses kata sandi",
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

	// Simpan ke Database
	if err := database.DB.Create(&newUser).Error; err != nil {
		// Asumsi error karena duplikasi (Username sudah terdaftar)
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat pengguna. Username mungkin sudah digunakan.",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Pengguna berhasil ditambahkan!",
		"data":    newUser, // Password otomatis tersembunyi berkat json:"-" di struct model
	})
}

// ========================================================
// 3. UPDATE USER (Parsial)
// ========================================================
func ApiUpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User

	// 1. Pastikan pengguna ada
	if err := database.DB.First(&user, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Pengguna tidak ditemukan",
		})
	}

	// 2. Parsing request JSON ke dalam bentuk Map yang fleksibel
	var input map[string]interface{}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid. Pastikan mengirim format JSON.",
		})
	}

	// 3. Siapkan keranjang data untuk GORM (kolom di-mapping sesuai database)
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

	// 4. Penanganan Khusus: Hashing Password jika pengguna ingin mengubahnya
	if val, ok := input["password"]; ok {
		passwordStr, isString := val.(string)

		// Hanya proses jika password berupa string dan tidak kosong
		if isString && passwordStr != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordStr), bcrypt.DefaultCost)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Gagal memproses kata sandi baru",
				})
			}
			updateData["password"] = string(hashedPassword)
		}
	}

	// 5. Eksekusi Update ke Database HANYA jika ada data di keranjang
	if len(updateData) > 0 {
		if err := database.DB.Model(&user).Updates(updateData).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal memperbarui data pengguna. Username mungkin sudah terpakai.",
			})
		}
	}

	// 6. Muat ulang data terbaru (agar response tidak mengembalikan data usang)
	database.DB.First(&user, id)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Data pengguna berhasil diperbarui!",
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
			"message": "Pengguna tidak ditemukan",
		})
	}

	// Hapus pengguna
	// Catatan: Jika di models.User kamu menambahkan gorm.DeletedAt,
	// maka ini otomatis menjadi Soft Delete seperti pada Posts.
	if err := database.DB.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus pengguna",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Pengguna berhasil dihapus",
	})
}
