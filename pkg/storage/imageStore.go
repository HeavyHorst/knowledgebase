package storage

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"strings"
	"time"

	"io/ioutil"

	"github.com/boltdb/bolt"
	"github.com/timshannon/bolthold"
)

type ImageStore interface {
	Get(hash string) []byte
	Insert(uri string) (string, error)
}

type defaultImageStore struct {
	store *bolthold.Store
}

func newDefaultImageStore(store *bolthold.Store) (defaultImageStore, error) {
	db := store.Bolt()
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("ImageStore"))
		if err != nil {
			return err
		}
		return nil
	})
	return defaultImageStore{store}, err
}

func (dis defaultImageStore) Get(hash string) []byte {
	var data []byte
	db := dis.store.Bolt()
	db.View(func(tx *bolt.Tx) error {
		data = tx.Bucket([]byte("ImageStore")).Get([]byte(hash))
		return nil
	})
	return data
}

func (dis defaultImageStore) Insert(uri string) (string, error) {
	if !strings.HasPrefix(uri, "http") {
		return "", nil
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(uri)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	image, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	hash := fmt.Sprintf("%x", sha1.Sum(image))

	db := dis.store.Bolt()
	err = db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("ImageStore")).Put([]byte(hash), image)
	})

	return hash, err
}
