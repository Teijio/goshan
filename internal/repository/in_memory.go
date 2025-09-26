package repository

import (
	"errors"
	"sync"
)

type URLRepository struct {
	store map[string]string
	mutex sync.RWMutex
}


func NewURLRepository() *URLRepository {
	return &URLRepository{
		store: make(map[string]string),
		mutex: sync.RWMutex{},
	}
}

func (r *URLRepository) Save(short, original string) error {
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

func (r *URLRepository) Get(short string) (string, error) {
    r.mutex.RLock()
	original, ok := r.store[short]
    r.mutex.RUnlock()

	if !ok {
		return "", errors.New("can't find full url by id")
	}

	return original, nil
}
