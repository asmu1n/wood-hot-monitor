package database

type DB struct{}

func New(dataDir string) (*DB, error) {
	return &DB{}, nil
}

func (db *DB) Close() error {
	return nil
}
