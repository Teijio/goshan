package repository

import (
	"errors"
	"sync"

	"github.com/Teijio/goshan/internal/models"
)

type InMemoreRepository struct {
	store map[string]models.ShortURL
	mutex sync.RWMutex
}

func NewInMemoreRepository() *InMemoreRepository {
	return &InMemoreRepository{
		store: make(map[string]models.ShortURL),
		mutex: sync.RWMutex{},
	}
}

func (r *InMemoreRepository) Save(shortURL models.ShortURL) error {
	r.mutex.RLock()
	_, ok := r.store[shortURL.ID]
	r.mutex.RUnlock()

	if ok {
		return errors.New("not unique id")
	}

	r.mutex.Lock()
	r.store[shortURL.ID] = shortURL
	r.mutex.Unlock()
	return nil
}

func (r *InMemoreRepository) Get(short string) (models.ShortURL, error) {
	r.mutex.RLock()
	original, ok := r.store[short]
	r.mutex.RUnlock()

	if !ok {
		return models.ShortURL{}, errors.New("can't find full url by id")
	}

	return original, nil
}

func (r *InMemoreRepository) Check() error {
	return nil
}
