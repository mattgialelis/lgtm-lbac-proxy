package main

import (
	"encoding/json"
	"os"

	"github.com/dgraph-io/badger"
)

//	{
//		"Name": "exampleName",
//		"Token": "exampleToken",
//		"TenantIds": ["tenant1", "tenant2"],
//		"AllowedLabels": {
//			"label1": [{"Type": "LABEL_EQ", "Name": "name1", "Value": "value1"}],
//			"label2": [{"Type": "LABEL_NEQ", "Name": "name2", "Value": "value2"}]
//		}
//	}

type KeyData struct {
	Name          string
	Token         string
	TenantIds     []string
	AllowedLabels map[string]string
}

type Store struct {
	db *badger.DB
}

func NewStore(path string) (*Store, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return nil, err
		}
	}

	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Add(key string, data KeyData) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return err
		}
		err = txn.Set([]byte(key), dataBytes)
		return err
	})
	return err
}

func (s *Store) Get(key string) (KeyData, bool, error) {
	var data KeyData
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}

		err = item.Value(func(val []byte) error {
			err = json.Unmarshal(val, &data)
			return err
		})
		return err
	})

	if err != nil {
		return KeyData{}, false, err
	}

	return data, true, nil
}

func (s *Store) GetAll() ([]KeyData, error) {
	var tokens []KeyData
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var token KeyData

				err := json.Unmarshal(val, &token)
				if err != nil {
					return err
				}
				tokens = append(tokens, token)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
