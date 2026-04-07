package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

type BadgerStore struct {
	db *badger.DB
}

func NewBadgerStore(path string) (*BadgerStore, error) {
	opts := badger.DefaultOptions(path).WithLogger(nil)
	opts.IndexCacheSize = 64 << 20
	opts.BaseTableSize = 2 << 20
	opts.ValueThreshold = 1024
	// opts.Compression = badger.
	opts.ValueLogFileSize = 16 << 20
	// opts.ValueLogLoadingMode = options.FileIO
	opts = opts.WithInMemory(false)
	opts = opts.WithSyncWrites(false)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &BadgerStore{db: db}, nil
}

func (b *BadgerStore) SaveDocumentWithTokens(ctx context.Context, id string, data []byte, tokens []string) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		docKey := []byte(fmt.Sprintf("doc:%s", id))
		return txn.Set(docKey, data)
	})
	if err != nil {
		return err
	}

	batchSize := 1000
	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		err := b.db.Update(func(txn *badger.Txn) error {
			for _, token := range tokens[i:end] {
				t := strings.ToLower(strings.TrimSpace(token))
				if t == "" {
					continue
				}
				idxKey := []byte(fmt.Sprintf("idx:%s:%s", t, id))
				if err := txn.Set(idxKey, nil); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BadgerStore) GetDocumentName(id string) (string, error) {
	var docName string
	err := b.db.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("doc:%s", id))
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			var doc struct {
				Title string `json:"Title"`
			}
			if err := json.Unmarshal(val, &doc); err != nil {
				return err
			}
			docName = doc.Title
			return nil
		})
	})
	return docName, err
}

func (b *BadgerStore) UpdateInvertedIndex(ctx context.Context, token string, docID string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("idx:%s:%s", token, docID))
		return txn.Set(key, nil)
	})
}

func (b *BadgerStore) GetDocIDsByToken(ctx context.Context, token string) ([]string, error) {
	var docIDs []string
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(fmt.Sprintf("idx:%s:", token))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := string(it.Item().Key())
			parts := strings.Split(key, ":")
			if len(parts) >= 3 {
				docIDs = append(docIDs, parts[len(parts)-1])
			}
		}
		return nil
	})
	return docIDs, err
}

func (b *BadgerStore) GetDocument(ctx context.Context, id string) ([]byte, error) {
	var data []byte
	err := b.db.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("doc:%s", id))
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			data = append([]byte{}, val...)
			return nil
		})
	})
	return data, err
}

func (b *BadgerStore) Close() error {
	return b.db.Close()
}
