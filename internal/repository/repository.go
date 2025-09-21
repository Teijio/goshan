package repository

type URLRepository struct {
    store map[string]string
}

func NewURLRepository() *URLRepository {
    return &URLRepository{
        store: make(map[string]string),
    }
}

func (r *URLRepository) Save(short, original string) {
    r.store[short] = original
}

func (r *URLRepository) Get(short string) (string, bool) {
    original, ok := r.store[short]
    return original, ok
}
