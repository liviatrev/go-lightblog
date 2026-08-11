package utils

import (
	"fmt"
	"go-lightblog/models"
	"net/url"
	"strings"
)

// GenerateBreadcrumbs creates a list of breadcrumb items for a given post.
func GenerateBreadcrumbs(post models.Post) []models.BreadcrumbItem {
	var breadcrumbs []models.BreadcrumbItem

	// Home Breadcrumb
	breadcrumbs = append(breadcrumbs, models.BreadcrumbItem{
		Name: "Home",
		URL:  "/",
	})

	// Category Breadcrumb (if post has a category)
	if post.Category.ID != 0 {
		breadcrumbs = append(breadcrumbs, models.BreadcrumbItem{
			Name: post.Category.Name,
			URL:  fmt.Sprintf("/category/%s", post.Category.Slug),
		})
	}

	// Post Title Breadcrumb
	breadcrumbs = append(breadcrumbs, models.BreadcrumbItem{
		Name: post.Title,
		URL:  fmt.Sprintf("/post/%s", post.Slug),
	})

	return breadcrumbs
}

// GenerateBreadcrumbListSchema generates the JSON-LD BreadcrumbList schema.
func GenerateBreadcrumbListSchema(breadcrumbs []models.BreadcrumbItem, baseURL string) models.BreadcrumbList {
	var itemList []models.BreadcrumbListItem
	for i, item := range breadcrumbs {
		absURL, _ := url.JoinPath(baseURL, strings.TrimPrefix(item.URL, "/"))
		itemList = append(itemList, models.BreadcrumbListItem{
			Type:    "ListItem",
			Position: i + 1,
			Name:    item.Name,
			Item:    absURL,
		})
	}

	return models.BreadcrumbList{
		Context:  "https://schema.org",
		Type:     "BreadcrumbList",
		ItemList: itemList,
	}
}
