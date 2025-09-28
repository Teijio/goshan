package repository

import (
	"errors"
	"sync"
)

type InMemoreRepository struct {
	store map[string]string
	mutex sync.RWMutex
}

func NewInMemoreRepository() *InMemoreRepository {
	return &InMemoreRepository{
		store: make(map[string]string),
		mutex: sync.RWMutex{},
	}
}

func (r *InMemoreRepository) Save(short, original string) error {
	r.mutex.RLock()
	_, ok := r.store[short]
	r.mutex.RUnlock()

	if ok {
		return errors.New("not unique id")
	}

	r.mutex.Lock()
	r.store[short] = original
	r.mutex.Unlock()
	return nil
}

func (r *InMemoreRepository) Get(short string) (string, error) {
	r.mutex.RLock()
	original, ok := r.store[short]
	r.mutex.RUnlock()

	if !ok {
		return "", errors.New("can't find full url by id")
	}

	return original, nil
}

func (r *InMemoreRepository) Check() error {
	return nil
}
