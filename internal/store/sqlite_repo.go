package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mattn/go-sqlite3"

	"PasswordManager/internal/db"
	"PasswordManager/domain"
)

type SQLiteRepository struct {
	queries *db.Queries
}

//create new wrapper
func NewSQLiteRepo(q *db.Queries) domain.Repository {
	return &SQLiteRepository{
		queries:	q,
	}
}

func (r *SQLiteRepository) CreateEntry(ctx context.Context, websiteURL, username string, encryptesPass []byte) (*domain.Entry, error) {
	entry, err := r.queries.CreateEntry(ctx, db.CreateEntryParams{
		WebsiteUrl: websiteURL,
		Username: username,
		EncrypedPassword: encryptesPass,
	})
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return nil, domain.ErrDuplicateEntry
			}
		}
	return nil, domain.ErrStorageFailure
	}
	return &domain.Entry{
		ID:			entry.ID,
		WebsiteURL: entry.WebsiteUrl,
		Username: 	entry.Username,
		EncryptedPassword: entry.EncrypedPassword,
	}, nil
}

func (r *SQLiteRepository) GetEntry(ctx context.Context, websiteurl, username string) (*domain.Entry, error) {
	entry, err := r.queries.GetEntry(ctx, db.GetEntryParams{
		WebsiteUrl:  websiteurl,
		Username: username,
	})

	if err != nil {
		// Catch standard "no results found" error
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrEntryNotFound
		}
		return nil, domain.ErrStorageFailure
	}

	return &domain.Entry{
		ID:                entry.ID,
		WebsiteURL:        entry.WebsiteUrl,
		Username:          entry.Username,
		EncryptedPassword: entry.EncrypedPassword,
	}, nil
}