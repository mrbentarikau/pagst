package rss

import (
	"time"
)

type RSSFeedTemplate struct {
	Title           string           `json:"title,omitempty"`
	Description     string           `json:"description,omitempty"`
	Link            string           `json:"link,omitempty"`
	FeedLink        string           `json:"feedLink,omitempty"`
	Links           []string         `json:"links,omitempty"`
	Updated         string           `json:"updated,omitempty"`
	UpdatedParsed   *time.Time       `json:"updatedParsed,omitempty"`
	Published       string           `json:"published,omitempty"`
	PublishedParsed *time.Time       `json:"publishedParsed,omitempty"`
	Author          *Person          `json:"author,omitempty"` // Deprecated: Use feed.Authors instead
	Authors         []*Person        `json:"authors,omitempty"`
	Language        string           `json:"language,omitempty"`
	Image           *Image           `json:"imasge,omitempty"`
	Copyright       string           `json:"copyright,omitempty"`
	Generator       string           `json:"generator,omitempty"`
	Categories      []string         `json:"categories,omitempty"`
	Item            *Item            `json:"item"`
	ItemsFiltered   RSSItemsFiltered `json:"items_filtered"`
	Items           []*Item          `json:"items"`
	FeedType        string           `json:"feedType"`
	FeedVersion     string           `json:"feedVersion"`
}

type Item struct {
	Title           string     `json:"title,omitempty"`
	Description     string     `json:"description,omitempty"`
	Content         string     `json:"content,omitempty"`
	Link            string     `json:"link,omitempty"`
	Links           []string   `json:"links,omitempty"`
	Updated         string     `json:"updated,omitempty"`
	UpdatedParsed   *time.Time `json:"updatedParsed,omitempty"`
	Published       string     `json:"published,omitempty"`
	PublishedParsed *time.Time `json:"publishedParsed,omitempty"`
	Author          *Person    `json:"author,omitempty"` // Deprecated: Use item.Authors instead
	Authors         []*Person  `json:"authors,omitempty"`
	GUID            string     `json:"guid,omitempty"`
	Image           *Image     `json:"image,omitempty"`
	Categories      []string   `json:"categories,omitempty"`
}

type RSSItemsFiltered []struct {
	Item
}

type Person struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type Image struct {
	URL   string `json:"url,omitempty"`
	Title string `json:"title,omitempty"`
}
