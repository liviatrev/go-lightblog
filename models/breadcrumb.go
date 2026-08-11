package models

// BreadcrumbItem represents a single item in the breadcrumb navigation.
type BreadcrumbItem struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// BreadcrumbList represents the full list of breadcrumbs for JSON-LD schema.
type BreadcrumbList struct {
	Context  string           `json:"@context"`
	Type     string           `json:"@type"`
	ItemList []BreadcrumbListItem `json:"itemListElement"`
}

// BreadcrumbListItem represents an item within the BreadcrumbList for JSON-LD.
type BreadcrumbListItem struct {
	Type    string `json:"@type"`	
	Position int    `json:"position"`
	Name    string `json:"name"`
	Item    string `json:"item"`
}
