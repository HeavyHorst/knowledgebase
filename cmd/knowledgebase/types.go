package main

import (
	"io"

	"github.com/HeavyHorst/knowledgebase/pkg/models"
)

type contextKey string

var (
	contextKeyArticle     = contextKey("article")
	contextKeyCategory    = contextKey("category")
	contextKeyCurrentUser = contextKey("currentUser")
	contextKeyUser        = contextKey("user")
)

type Backuper interface {
	Backup(w io.Writer) error
}

type ArticleGetter interface {
	GetArticle(id string) (models.Article, error)
}

type ArticleHistoryGetter interface {
	GetArticleHistory(artID string) ([]models.ArticleHistoryEntry, error)
}

type ArticleLister interface {
	ListArticles(limit, offset int) ([]models.Article, error)
	ListArticlesForCategory(catID string) ([]models.Article, error)
}

type ArticleCreator interface {
	CreateArticle(art models.Article, author models.User) error
}

type ArticleUpdater interface {
	UpdateArticle(art models.Article, author models.User) error
}

type ArticleDeleter interface {
	DeleteArticle(models.Article) error
}

type ArticleSearcher interface {
	SearchArticles(query string) ([]models.Article, error)
}

type CategoryGetter interface {
	GetCategory(id string) (models.Category, error)
}

type CategoryLister interface {
	ListCategories() ([]models.Category, error)
	ListBaseCategories() ([]models.Category, error)
	ListCategoriesForCategory(catID string) ([]models.Category, error)
}

type CategoryCreator interface {
	CreateCategory(models.Category) error
}

type CategoryUpdater interface {
	UpdateCategory(models.Category) error
}

type CategoryDeleter interface {
	DeleteCategory(models.Category) error
}

type CategorySearcher interface {
	SearchCategories(query string) ([]models.Category, error)
}

type ImageGetter interface {
	GetImage(hash string) []byte
}

type Authenticator interface {
	Authenticate(username, password string) (*models.User, error)
}

type UserLister interface {
	ListUsers() ([]models.User, error)
}

type UserCreator interface {
	CreateUser(user models.User, password string) error
}

type UserGetter interface {
	GetUser(name string) (models.User, error)
}

type UserDeleter interface {
	DeleteUser(models.User) error
}

type UserUpdater interface {
	UpdateUser(user models.User, password string) error
}
