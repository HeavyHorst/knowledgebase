package storage

import (
	"fmt"
	"time"

	"github.com/HeavyHorst/knowledgebase/pkg/models"
	"github.com/blevesearch/bleve"
	"github.com/pkg/errors"
	"github.com/timshannon/bolthold"
)

var categoryIndex bleve.Index
var ErrSameCategory = errors.New("a category can not be assigned to itself")

func init() {
	var err error
	cmapping := bleve.NewIndexMapping()
	categoryIndex, err = bleve.Open("data/category.bleve")
	if err != nil {
		if err == bleve.ErrorIndexPathDoesNotExist {
			categoryIndex, err = bleve.New("data/category.bleve", cmapping)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}

type CategoryStore struct {
	store      *bolthold.Store
	ImageStore ImageStore
}

func newCategoryStore(store *bolthold.Store, is ImageStore) (*CategoryStore, error) {
	return &CategoryStore{
		store:      store,
		ImageStore: is,
	}, nil
}

func (b *CategoryStore) GetCategory(id string) (models.Category, error) {
	var cat models.Category
	if err := b.store.Get(id, &cat); err != nil {
		return models.Category{}, errors.Wrapf(err, "couldn't get category %s", id)
	}

	return cat, nil
}

func (b *CategoryStore) ListCategories() ([]models.Category, error) {
	var result []models.Category
	err := b.store.Find(&result, nil)
	return result, errors.Wrap(err, "couldn't get category list")
}

func (b *CategoryStore) ListCategoriesForCategory(catID string) ([]models.Category, error) {
	var result []models.Category
	err := b.store.Find(&result, bolthold.Where("Category").Eq(catID))
	return result, errors.Wrapf(err, "couldn't get categories for category %s", catID)
}

func (b *CategoryStore) upsertCategory(cat models.Category, typ insertType) error {
	hash, err := b.ImageStore.Insert(cat.Image)
	if err != nil {
		return errors.Wrapf(err, "couldn't update %s in the image store", cat.ID)
	}
	if hash != "" {
		cat.Image = "/image/" + hash
	}

	cat.LastModified = time.Now()

	if cat.ID == cat.Category {
		return ErrSameCategory
	}

	switch typ {
	case insertTypeCreate:
		err = b.store.Insert(cat.ID, cat)
	case insertTypeUpdate:
		err = b.store.Update(cat.ID, cat)
	}

	if err != nil {
		return errors.Wrapf(err, "couldn't insert %s into the store", cat.ID)
	}

	err = categoryIndex.Index(cat.ID, cat)
	if err != nil {
		return errors.Wrapf(err, "couldn't add %s to the fulltext index", cat.ID)
	}

	return nil
}

func (b *CategoryStore) CreateCategory(cat models.Category) error {
	return b.upsertCategory(cat, insertTypeCreate)
}

func (b *CategoryStore) UpdateCategory(cat models.Category) error {
	return b.upsertCategory(cat, insertTypeUpdate)
}

func (b *CategoryStore) DeleteCategory(cat models.Category) error {
	// delete from store
	err := b.store.Delete(cat.ID, cat)
	if err != nil {
		return errors.Wrapf(err, "couldn't delete %s from the store", cat.ID)
	}

	// delete from index
	err = categoryIndex.Delete(cat.ID)
	if err != nil {
		return errors.Wrapf(err, "couldn't delete %s from the fulltext index", cat.ID)
	}

	return nil
}

func (b *CategoryStore) SearchCategories(q string) ([]models.Category, error) {
	query := bleve.NewQueryStringQuery(q)
	search := bleve.NewSearchRequestOptions(query, 150, 0, false)
	search.Highlight = bleve.NewHighlightWithStyle("html")
	searchResults, err := categoryIndex.Search(search)
	if err != nil {
		return nil, errors.Wrap(err, "bleve index search failed")
	}

	var lastErrs error
	cats := make([]models.Category, 0, len(searchResults.Hits))

	for _, v := range searchResults.Hits {
		var cat models.Category
		err := b.store.Get(v.ID, &cat)
		if err != nil {
			lastErrs = errors.Wrap(lastErrs, fmt.Sprintf("%s: couldn't get article %s", err.Error(), v.ID))
		} else {
			cat.Fragments = v.Fragments
			cats = append(cats, cat)
		}
	}
	return cats, lastErrs
}
