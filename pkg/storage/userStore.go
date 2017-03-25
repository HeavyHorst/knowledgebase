package storage

import (
	"os"
	"strings"

	"github.com/HeavyHorst/knowledgebase/pkg/models"
	"github.com/pkg/errors"
	"github.com/timshannon/bolthold"
	"golang.org/x/crypto/bcrypt"
)

type insertType string

var (
	insertTypeUpdate = insertType("update")
	insertTypeCreate = insertType("create")
)

type UserStore struct {
	store      *bolthold.Store
	ImageStore ImageStore
}

func newUserStore(store *bolthold.Store, is ImageStore) (*UserStore, error) {
	var err error
	un := os.Getenv("KB_USER")
	pw := os.Getenv("KB_PASSWORD")

	userstore := &UserStore{
		store:      store,
		ImageStore: is,
	}

	if un != "" {
		err = userstore.CreateUser(models.User{
			UserInfo: models.UserInfo{
				Username: un,
			},
			IsAdmin: true,
		}, pw)
	}

	return userstore, err
}

func (b *UserStore) GetUser(name string) (models.User, error) {
	var user models.User
	if err := b.store.Get(name, &user); err != nil {
		return models.User{}, errors.Wrapf(err, "couldn't get user %s", name)
	}
	return user, nil
}

func (b *UserStore) upsertUser(user models.User, typ insertType, password string) error {
	if strings.TrimSpace(user.Username) != "" {

		// calculate the password hash if we create a new user or if we want to updated the password.
		// We leave the password as is if the password equals "" and the type is insertTypeUpdate.
		if typ == insertTypeCreate || (typ == insertTypeUpdate && password != "") {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
			if err != nil {
				return errors.Wrapf(err, "couldn't calculate the password hash")
			}
			user.Password = string(hashedPassword)
		} else {
			current, err := b.GetUser(user.Username)
			if err != nil {
				return errors.New("no password provided and couldn't get current one")
			}
			user.Password = current.Password
		}

		hash, err := b.ImageStore.Insert(user.Image)
		if err != nil {
			return errors.Wrapf(err, "couldn't insert %s into the image store", user.Image)
		}
		if hash != "" {
			user.Image = "/image/" + hash
		}

		switch typ {
		case insertTypeCreate:
			err = b.store.Insert(user.Username, user)
		case insertTypeUpdate:
			err = b.store.Update(user.Username, user)
		}

		if err != nil {
			return errors.Wrapf(err, "couldn't insert %s into the store", user.Username)
		}
	} else {
		return errors.New("username can't be empty")
	}

	return nil
}

func (b *UserStore) CreateUser(user models.User, password string) error {
	return b.upsertUser(user, insertTypeCreate, password)
}

func (b *UserStore) UpdateUser(user models.User, password string) error {
	return b.upsertUser(user, insertTypeUpdate, password)
}

func (b *UserStore) DeleteUser(user models.User) error {
	// delete from store
	err := b.store.Delete(user.Username, user)
	if err != nil {
		return errors.Wrapf(err, "couldn't delete %s from the store", user.Username)
	}
	return nil
}

func (b *UserStore) Authenticate(username, password string) (*models.User, error) {
	user, err := b.GetUser(username)
	if err != nil {
		return nil, err
	}

	if user.Username == username && bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil {
		return &user, nil
	}

	return nil, nil
}

func (b *UserStore) ListUsers() ([]models.User, error) {
	var result []models.User
	err := b.store.Find(&result, nil)

	for k := range result {
		result[k].Password = ""
	}

	return result, errors.Wrap(err, "couldn't get user list")
}
