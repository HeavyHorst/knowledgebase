package models

import (
	"time"

	"github.com/blevesearch/bleve/search"
)

type Article struct {
	ID       string
	Category string `boltholdIndex:"Category" json:"category"`
	Title    string `json:"title"`
	Short    string `json:"short"`
	Article  string `json:"article,omitempty"`

	LastModified time.Time `json:"last_modified"`
	Tags         []string  `json:"tags,omitempty"`

	// for search result highlighting
	Fragments search.FieldFragmentMap `json:"fragments,omitempty"`
	Authors   []UserInfo              `json:"authors"`
}

type Category struct {
	ID           string
	Category     string    `boltholdIndex:"Category" json:"category"`
	Image        string    `json:"image"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	LastModified time.Time `json:"last_modified"`

	// for search result highlighting
	Fragments search.FieldFragmentMap `json:"fragments,omitempty" msg:"-"`
}

type User struct {
	IsAdmin  bool   `json:"is_admin,omitempty"`
	Password string `json:"password,omitempty"`
	UserInfo `msgpack:",inline"`
}

type UserInfo struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Image     string `json:"image,omitempty"`
}

type ArticleHistoryEntry struct {
	Timestamp  time.Time
	ModifiedBy string
	ArticleID  string `boltholdIndex:"Article" json:"-"`
}
