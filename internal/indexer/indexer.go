package indexer

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/Pix3l-Guy21/search-engine/internal/pipeline"
	"github.com/Pix3l-Guy21/search-engine/internal/store"
)

type Indexer struct {
	mu      sync.RWMutex
	index   map[string][]string
	storage store.Storage
}

func NewIndexer(s store.Storage) *Indexer {
	return &Indexer{
		index:   make(map[string][]string),
		storage: s,
	}
}

func (idx *Indexer) IndexDocument(ctx context.Context, doc pipeline.Document) error {
	docBytes, err := json.Marshal(doc)
	if err != nil {
		return err // Handle the error if serialization fails
	}
	return idx.storage.SaveDocumentWithTokens(ctx, doc.ID, docBytes, doc.Tokens)
}

func (idx *Indexer) Search(ctx context.Context, term string) ([]string, error) {
	cleanTerm := strings.ToLower(strings.TrimSpace(term))
	return idx.storage.GetDocIDsByToken(ctx, cleanTerm)
}
