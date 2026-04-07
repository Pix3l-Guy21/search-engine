package store

import (
	"context"
)

type Storage interface {
	SaveDocumentWithTokens(ctx context.Context, id string, data []byte, token []string) error
	GetDocument(ctx context.Context, id string) ([]byte, error)
	UpdateInvertedIndex(ctx context.Context, token string, docId string) error
	GetDocIDsByToken(ctx context.Context, token string) ([]string, error)
	GetDocumentName(id string) (string, error)
	Close() error
}
