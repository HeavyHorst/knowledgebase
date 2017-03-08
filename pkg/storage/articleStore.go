package storage

import (
	"fmt"
	"time"

	"github.com/HeavyHorst/sunoKB/pkg/models"
	"github.com/blevesearch/bleve"
	"github.com/pkg/errors"
	"github.com/russross/blackfriday"
	"github.com/timshannon/bolthold"
)

var articleIndex bleve.Index

func init() {
	var err error
	amapping := bleve.NewIndexMapping()
	articleIndex, err = bleve.Open("data/article.bleve")
	if err != nil {
		if err == bleve.ErrorIndexPathDoesNotExist {
			articleIndex, err = bleve.New("data/article.bleve", amapping)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}

type ArticleStore struct {
	store *bolthold.Store
}

func newArticleStore(store *bolthold.Store) (*ArticleStore, error) {
	return &ArticleStore{
		store: store,
	}, nil
}

func (b *ArticleStore) GetArticle(id string) (models.Article, error) {
	var art models.Article
	if err := b.store.Get(id, &art); err != nil {
		return models.Article{}, errors.Wrapf(err, "couldn't get article %s", id)
	}

	return art, nil
}

func (b *ArticleStore) ListArticles(limit, offset int) ([]models.Article, error) {
	var result []models.Article
	err := b.store.Find(&result, bolthold.Where(bolthold.Key).Ne("").Skip(offset).Limit(limit))

	// we don't want the complete article in the listing
	for k := range result {
		result[k].Article = ""
	}

	return result, errors.Wrap(err, "couldn't get list of articles")
}

func (b *ArticleStore) ListArticlesForCategory(catID string) ([]models.Article, error) {
	var result []models.Article
	err := b.store.Find(&result, bolthold.Where("Category").Eq(catID))

	// we don't want the complete article in the listing
	for k := range result {
		result[k].Article = ""
	}

	return result, errors.Wrapf(err, "couldn't get articles for category %s", catID)
}

func (b *ArticleStore) upsertArticle(art models.Article, typ insertType) error {
	var err error
	art.LastModified = time.Now()

	switch typ {
	case insertTypeCreate:
		err = b.store.Insert(art.ID, art)
	case insertTypeUpdate:
		err = b.store.Update(art.ID, art)
	}

	if err != nil {
		return errors.Wrapf(err, "couldn't insert %s into the store", art.ID)
	}

	r := blackfriday.MarkdownCommon([]byte(art.Article))
	art.Article = htmlToText(r)

	err = articleIndex.Index(art.ID, art)
	if err != nil {
		return errors.Wrapf(err, "couldn't add %s to the fulltext index", art.ID)
	}

	return nil
}

func (b *ArticleStore) CreateArticle(art models.Article) error {
	return b.upsertArticle(art, insertTypeCreate)
}

func (b *ArticleStore) UpdateArticle(art models.Article) error {
	return b.upsertArticle(art, insertTypeUpdate)
}

func (b *ArticleStore) DeleteArticle(art models.Article) error {
	// delete from store
	err := b.store.Delete(art.ID, art)
	if err != nil {
		return errors.Wrapf(err, "couldn't delete %s from the store", art.ID)
	}

	// delete from index
	err = articleIndex.Delete(art.ID)
	if err != nil {
		return errors.Wrapf(err, "couldn't delete %s from the fulltext index", art.ID)
	}

	return nil
}

func (b *ArticleStore) SearchArticles(q string) ([]models.Article, error) {
	query := bleve.NewQueryStringQuery(q)
	search := bleve.NewSearchRequestOptions(query, 150, 0, false)
	search.Highlight = bleve.NewHighlightWithStyle("html")
	searchResults, err := articleIndex.Search(search)
	if err != nil {
		return nil, errors.Wrap(err, "bleve index search failed")
	}

	var lastErrs error
	arts := make([]models.Article, 0, len(searchResults.Hits))

	for _, v := range searchResults.Hits {
		var art models.Article
		err := b.store.Get(v.ID, &art)
		if err != nil {
			lastErrs = errors.Wrap(lastErrs, fmt.Sprintf("%s: couldn't get article %s", err.Error(), v.ID))
		} else {
			art.Article = ""
			art.Fragments = v.Fragments
			arts = append(arts, art)
		}
	}
	return arts, lastErrs
}
