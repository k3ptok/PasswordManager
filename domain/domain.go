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
	CreateEntry(ctx context.Context, websiteURL, username string, encryptedPass []byte) (*Entry, error)
	GetEntry(ctx context.Context, websiteURL, username string) (*Entry, error)
	ListEntries(ctx context.Context) ([]Entry, error)
	UpdateEntry(ctx context.Context, websiteURL, username string, encryptedPass []byte) error
	DeleteEntry(ctx context.Context, website_url, username string) error
	SearchEntries(ctx context.Context, keyword string) ([]Entry, error)
	ImportEntry(ctx context.Context, website_url, username string, encryptedPass []byte) error

	SetConfig(ctx context.Context,key string, value []byte) error
	GetConfig(ctx context.Context, key string) ([]byte, error)

}

var (
	ErrEntryNotFound   = errors.New("password entry not found")
    ErrDuplicateEntry  = errors.New("an entry for this website and username already exists")
    ErrStorageFailure  = errors.New("internal storage failure")
)