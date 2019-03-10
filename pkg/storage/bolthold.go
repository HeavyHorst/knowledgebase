package storage

import (
	"io"

	msgpack "gopkg.in/vmihailenco/msgpack.v2"

	"github.com/HeavyHorst/knowledgebase/pkg/models"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/timshannon/bolthold"
)

// custom msgpack decoding function for bolthold (faster than gobs)
func dec(data []byte, value interface{}) error {
	return msgpack.Unmarshal(data, value)
}

// custom msgpack encoding function for bolthold (faster than gobs)
func enc(value interface{}) ([]byte, error) {
	return msgpack.Marshal(value)
}

type BoltHoldClient struct {
	store      *bolthold.Store
	ImageStore ImageStore
	*ArticleStore
	*CategoryStore
	*UserStore
}

func NewBoltHoldClient(path string) (*BoltHoldClient, error) {
	store, err := bolthold.Open(path, 0666, &bolthold.Options{
		Encoder: bolthold.EncodeFunc(enc),
		Decoder: bolthold.DecodeFunc(dec),
	})
	if err != nil {
		return nil, err
	}

	// Garbage Collection for testing
	//store.DeleteMatching(&models.User{}, bolthold.Where("Username").Eq(""))

	is, err := newDefaultImageStore(store)
	if err != nil {
		return nil, err
	}

	us, err := newUserStore(store, is)
	if err != nil {
		return nil, err
	}

	as, err := newArticleStore(store, us, is)
	if err != nil {
		return nil, err
	}

	cs, err := newCategoryStore(store, is)
	if err != nil {
		return nil, err
	}

	return &BoltHoldClient{
		store:         store,
		ImageStore:    is,
		ArticleStore:  as,
		CategoryStore: cs,
		UserStore:     us,
	}, nil
}

func (b *BoltHoldClient) Backup(w io.Writer) error {
	return b.store.Bolt().View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(w)
		return err
	})
}

func (b *BoltHoldClient) GetImage(hash string) []byte {
	return b.ImageStore.Get(hash)
}

func (b *BoltHoldClient) UpdateUser(user models.User, password string) error {
	err := b.UserStore.UpdateUser(user, password)
	if err != nil {
		return err
	}

	err = b.ArticleStore.updateAllAuthors(user)
	return errors.Wrap(err, "couldn't update the article authors")
}
