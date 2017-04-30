package storage

import (
	"fmt"
	"sort"
	"time"

	"github.com/HeavyHorst/knowledgebase/pkg/models"
	"github.com/HeavyHorst/knowledgebase/pkg/ulid"
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

	return art, nil
}

func (b *ArticleStore) ListArticles(limit, offset int, sortBy string, reverse bool) ([]models.Article, int, error) {
	var result []models.Article

	err := b.store.Find(&result, nil)
	count := len(result)
	end := offset + limit
	if end >= count {
		end = count
	}

	switch sortBy {
	case "title":
		sort.Slice(result, func(i, j int) bool {
			return (result[i].Title < result[j].Title) != reverse
		})
	case "description":
		sort.Slice(result, func(i, j int) bool {
			return (result[i].Short < result[j].Short) != reverse
		})
	case "last_modified":
		sort.Slice(result, func(i, j int) bool {
			return (result[i].LastModified.After(result[j].LastModified)) != reverse
		})
	}

	subRes := result[offset:end]
	// we don't want the complete article in the listing
	for k := range subRes {
		subRes[k].Article = ""
		if subRes[k].Authors == nil {
			subRes[k].Authors = make([]models.UserInfo, 0)
		}
	}

	return subRes, count, errors.Wrap(err, "couldn't get list of articles")
}

func (b *ArticleStore) ListArticlesForCategory(catID string) ([]models.Article, error) {
	var result []models.Article
	err := b.store.Find(&result, bolthold.Where("Category").Eq(catID))

	sort.Slice(result, func(i, j int) bool {
		return result[i].LastModified.After(result[j].LastModified)
	})

	for k := range result {
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
