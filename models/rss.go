// models/rss.go
package models

import "encoding/xml"

// RSSItem represents a single <item> in the RSS 2.0 feed
type RSSItem struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	PubDate     string   `xml:"pubDate"`
	GUID        string   `xml:"guid"`
	Author      string   `xml:"author"`
	Categories  []string `xml:"category"`
}

// RSSChannel represents the <channel> element of an RSS 2.0 feed
type RSSChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language"`
	LastBuildDate string    `xml:"lastBuildDate"`
	AtomLink      AtomLink  `xml:"atom:link"`
	Items         []RSSItem `xml:"item"`
}

// RSS represents the root <rss> element of an RSS 2.0 feed
type RSS struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Xmlns   string     `xml:"xmlns:atom,attr"`
	Channel RSSChannel `xml:"channel"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}