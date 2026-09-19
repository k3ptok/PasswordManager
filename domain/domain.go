package domain

import (
	"context"
	"errors"
)

type Entry struct {
	ID					int64
	WebsiteURL			string
	Username			string
	EncryptedPassword	[]byte
}

type Repository interface {
	CreateEntry(ctx context.Context, websiteURL, username string, encryptesPass []byte) (*Entry, error)
	GetEntry(ctx context.Context, websiteURL, username string) (*Entry, error)
}

var (
	ErrEntryNotFound   = errors.New("password entry not found")
    ErrDuplicateEntry  = errors.New("an entry for this service and username already exists")
    ErrStorageFailure  = errors.New("internal storage failure")
)