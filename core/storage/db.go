package storage

import (
	"errors"

	badger "github.com/dgraph-io/badger"
	"github.com/rs/zerolog/log"
)

// Storage represents interface to talk to core store
type Storage struct {
	db *badger.DB
}

var intitialized bool = false

var (
	// ErrorStorageInitialized in case storage has already been initialized.
	ErrorStorageInitialized = errors.New("storage_already_initialized")
	//
	ErrorKeyNotFound = errors.New("key_not_found")
)

// Open creates a new storage instance
func Open() (*Storage, error) {
	if intitialized {
		return nil, ErrorStorageInitialized
	}
	db, err := badger.Open(badger.DefaultOptions("/tmp/raker"))
	if err != nil {
		intitialized = true
	}
	return &Storage{db: db}, err
}

func OpenForUnitTest(fileName string) (*Storage, error)  {
	if intitialized {
		return nil, ErrorStorageInitialized
	}
	db, err := badger.Open(badger.DefaultOptions("/tmp/" + fileName))
	if err != nil {
		intitialized = true
	}
	return &Storage{db: db}, err
}

// Get value by key from DB.
func (storage *Storage) Get(key string) ([]byte, error) {
	var output []byte
	err := storage.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		err = handleDbError(err)
		if err != nil {
			return err
		}
		err = item.Value(func(valueBytes []byte) error {
			output = valueBytes
			return nil
		})
		return err
	})
	return output, err
}

// GetAllByPrefix returns all values where key starts with a prefix.
func (storage *Storage) GetAllByPrefix(prefix string) ([][]byte, error) {
	methodName := "GetAllByPrefix"
	log.Debug().Msgf("%s: Finding all values for prefix - %s", methodName, prefix)
	outputBytes := make([][]byte, 4)
	err := storage.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefixBytes := []byte(prefix)
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			item := it.Item()
			log.Debug().Msgf("%s: Found key %s during iteration.", methodName, item.Key())
			err := item.Value(func(v []byte) error {
				outputBytes = append(outputBytes, v)
				return nil
			})
			err = handleDbError(err)
			if err != nil {
				log.Debug().Msgf("%s: Got error while fetching value for key %s", methodName, item.Key())
				return err
			}
		}
		return nil
	})
	return outputBytes, err
}

// Put will add or update new key in case it is already present.
func (storage *Storage) Put(key string, value []byte) error {
	err := storage.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), value)
		return err
	})
	return err
}

// Close the db instance
func (storage *Storage) Close() {
	storage.db.Close()
}

func (storage *Storage) CleanUpFoUnitTests() {
	storage.db.DropAll()
	storage.Close()
}

func handleDbError(err error) error {
	if err == badger.ErrKeyNotFound {
		return ErrorKeyNotFound
	}
	return err
}
