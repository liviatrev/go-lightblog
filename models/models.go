// models/models.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// User sekarang mendukung RBAC dan API Key
type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null" json:"-"`
	Name     string `gorm:"not null"`             // Nama Tampilan (Author)
	Role     string `gorm:"default:'editor'"`     // 'admin' atau 'editor'
	APIKey   string `gorm:"uniqueIndex;not null" json:"-"` // Untuk akses Headless API
	Posts    []Post `gorm:"foreignKey:AuthorID" json:"posts,omitempty"`  // Relasi One-to-Many ke Post

	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// Setting tetap sama, tapi akan menampung lebih banyak key (Disqus, Gemini API, ImageKit)
type Setting struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"not null"`
}

// Category (Satu post memiliki satu kategori utama)
type Category struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"not null"`
	Slug  string `gorm:"uniqueIndex;not null"`
	Posts []Post `json:"posts,omitempty"` // Relasi Has-Many
}

// Tag (Satu post bisa punya banyak tag, satu tag bisa dimiliki banyak post)
type Tag struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"not null"`
	Slug  string `gorm:"uniqueIndex;not null"`
	Posts []Post `gorm:"many2many:post_tags;" json:"posts,omitempty"` // Relasi Many-to-Many
}

// Post berevolusi dengan kolom SEO, Author, dan Relasi
type Post struct {
	ID              uint      `gorm:"primaryKey"`
	Title           string    `gorm:"not null"`
	Slug            string    `gorm:"uniqueIndex;not null"`
	Content         string    
	IsDraft         bool      
	CoverImage      string    // URL gambar dari CDN (ImageKit/S3)
	Type            string    `gorm:"type:varchar(20);default:'post'"`
	
	// Kolom Khusus SEO
	MetaTitle       string
	MetaDescription string
	TargetKeyword   string

	// Relasi ke User (Author)
	AuthorID        uint
	Author          User      `gorm:"foreignKey:AuthorID"`

	// Relasi ke Category
	CategoryID      uint
	Category        Category  `gorm:"foreignKey:CategoryID"`

	// Relasi ke Tag (Many-to-Many)
	Tags            []Tag     `gorm:"many2many:post_tags;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}