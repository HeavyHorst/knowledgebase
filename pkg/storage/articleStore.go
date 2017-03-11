package storage

import (
	"fmt"
	"time"

	"github.com/HeavyHorst/sunoKB/pkg/models"
	"github.com/HeavyHorst/sunoKB/pkg/ulid"
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
	store     *bolthold.Store
	userStore *UserStore
}

func newArticleStore(store *bolthold.Store, userStore *UserStore) (*ArticleStore, error) {
	return &ArticleStore{
		store:     store,
		userStore: userStore,
	}, nil
}

func (b *ArticleStore) updateAllAuthors(author models.User) error {
	return b.store.UpdateMatching(&models.Article{}, bolthold.Where("Authors").MatchFunc(func(ra *bolthold.RecordAccess) (bool, error) {
		record := ra.Record()
		article, ok := record.(*models.Article)
		if ok {
			for _, v := range article.Authors {
				if v.Username == author.Username {
					return true, nil
				}
			}
		}
		return false, nil

	}), func(record interface{}) error {
		update, ok := record.(*models.Article) // record will always be a pointer
		if !ok {
			return fmt.Errorf("Record isn't the correct type!  Wanted &models.Article, got %T", record)
		}

		for k := range update.Authors {
			if update.Authors[k].Username == author.Username {
				update.Authors[k] = author.UserInfo
			}
		}

		return nil
	})
}

func (b *ArticleStore) GetArticle(id string) (models.Article, error) {
	var art models.Article
	if err := b.store.Get(id, &art); err != nil {
		return art, errors.Wrapf(err, "couldn't get article %s", id)
	}

	if art.Authors == nil {
		art.Authors = make([]models.UserInfo, 0)
	}

	/*authors, err := b.getAuthorsForArticle(id, nil)
	if err != nil {
		return art, err
	}
	art.Authors = authors*/

	return art, nil
}

func (b *ArticleStore) ListArticles(limit, offset int) ([]models.Article, error) {
	var result []models.Article
	err := b.store.Find(&result, bolthold.Where(bolthold.Key).Ne("").Skip(offset).Limit(limit))
	//cache := make(map[string]models.UserInfo)

	// we don't want the complete article in the listing
	for k := range result {
		/*authors, err := b.getAuthorsForArticle(result[k].ID, cache)
		if err != nil {
			return nil, err
		}
		result[k].Authors = authors*/
		result[k].Article = ""
		if result[k].Authors == nil {
			result[k].Authors = make([]models.UserInfo, 0)
		}
	}

	return result, errors.Wrap(err, "couldn't get list of articles")
}

func (b *ArticleStore) ListArticlesForCategory(catID string) ([]models.Article, error) {
	var result []models.Article
	err := b.store.Find(&result, bolthold.Where("Category").Eq(catID))
	//cache := make(map[string]models.UserInfo)

	// we don't want the complete article in the listing
	for k := range result {
		/*authors, err := b.getAuthorsForArticle(result[k].ID, cache)
		if err != nil {
			return nil, err
		}
		result[k].Authors = authors*/
		result[k].Article = ""
		if result[k].Authors == nil {
			result[k].Authors = make([]models.UserInfo, 0)
		}
	}

	return result, errors.Wrapf(err, "couldn't get articles for category %s", catID)
}

func (b *ArticleStore) GetArticleHistory(artID string) ([]models.ArticleHistoryEntry, error) {
	var result []models.ArticleHistoryEntry
	err := b.store.Find(&result, bolthold.Where("ArticleID").Eq(artID))
	return result, errors.Wrapf(err, "couldn't get history for article %s", artID)
}

func (b *ArticleStore) upsertArticle(art models.Article, typ insertType, author models.User) error {
	var err error
	art.LastModified = time.Now()

	b.store.Insert(ulid.GetULID(), models.ArticleHistoryEntry{
		Timestamp:  art.LastModified,
		ModifiedBy: author.Username,
		ArticleID:  art.ID,
	})

	art.Authors = append(art.Authors, author.UserInfo)
	if len(art.Authors) >= 3 {
		art.Authors = art.Authors[len(art.Authors)-3:]
	}

	if len(art.Authors) >= 2 {
		a := art.Authors[len(art.Authors)-1]
		b := art.Authors[len(art.Authors)-2]

		if a.FirstName == b.FirstName && a.LastName == b.LastName {
			art.Authors = art.Authors[:len(art.Authors)-1]
		}
	}

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

func (b *ArticleStore) CreateArticle(art models.Article, author models.User) error {
	return b.upsertArticle(art, insertTypeCreate, author)
}

func (b *ArticleStore) UpdateArticle(art models.Article, author models.User) error {
	return b.upsertArticle(art, insertTypeUpdate, author)
}

func (b *ArticleStore) DeleteArticle(art models.Article) error {
	// delete from store
	err := b.store.Delete(art.ID, art)
	if err != nil {
		return errors.Wrapf(err, "couldn't delete %s from the store", art.ID)
	}

	// delete the article change history
	err = b.store.DeleteMatching(&models.ArticleHistoryEntry{}, bolthold.Where("ArticleID").Eq(art.ID))
	if err != nil {
		return errors.Wrapf(err, "couldn't delete the history for %s", art.ID)
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
		art, err := b.GetArticle(v.ID)
		if err != nil {
			lastErrs = errors.Wrap(lastErrs, err.Error())
		} else {
			art.Article = ""
			art.Fragments = v.Fragments
			arts = append(arts, art)
		}
	}
	return arts, lastErrs
}
