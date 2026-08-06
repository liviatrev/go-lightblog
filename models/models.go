// models/models.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// User now supports RBAC and API Key
type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null" json:"-"`
	Name     string `gorm:"not null"`             // Display Name (Author)
	Role     string `gorm:"default:'editor'"`     // 'admin' or 'editor'
	APIKey   string `gorm:"uniqueIndex;not null" json:"-"` // For Headless API access
	Posts    []Post `gorm:"foreignKey:AuthorID" json:"posts,omitempty"`  // One-to-Many relationship to Post

	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// Setting remains the same, but will hold more keys (Disqus, Gemini API, ImageKit)
type Setting struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"not null"`
}

// Category (One post has one main category)
type Category struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"not null"`
	Slug  string `gorm:"uniqueIndex;not null"`
	Posts []Post `json:"posts,omitempty"` // Has-Many relationship
}

// Tag (One post can have many tags, one tag can belong to many posts)
type Tag struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"not null"`
	Slug  string `gorm:"uniqueIndex;not null"`
	Posts []Post `gorm:"many2many:post_tags;" json:"posts,omitempty"` // Many-to-Many relationship
}

// Post evolved with SEO columns, Author, and Relationships
type Post struct {
	ID              uint      `gorm:"primaryKey"`
	Title           string    `gorm:"not null"`
	Slug            string    `gorm:"uniqueIndex;not null"`
	Content         string    
	IsDraft         bool      
	CoverImage      string    // Image URL from CDN (ImageKit/S3)
	Type            string    `gorm:"type:varchar(20);default:'post'"`
	
	// SEO-specific columns
	MetaTitle       string
	MetaDescription string
	TargetKeyword   string

	// Relationship to User (Author)
	AuthorID        uint
	Author          User      `gorm:"foreignKey:AuthorID"`

	// Relationship to Category
	CategoryID      uint
	Category        Category  `gorm:"foreignKey:CategoryID"`

	// Relationship to Tag (Many-to-Many)
	Tags            []Tag     `gorm:"many2many:post_tags;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
