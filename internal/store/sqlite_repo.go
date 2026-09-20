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

func (r *SQLiteRepository) CreateEntry(ctx context.Context, websiteURL, username string, encryptedPass []byte) (*domain.Entry, error) {
	entry, err := r.queries.CreateEntry(ctx, db.CreateEntryParams{
		WebsiteUrl: websiteURL,
		Username: username,
		EncryptedPassword: encryptedPass,
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
		EncryptedPassword: entry.EncryptedPassword,
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
		EncryptedPassword: entry.EncryptedPassword,
	}, nil
}

func (r *SQLiteRepository) ListEntries(ctx context.Context) ([]domain.Entry, error) {
	rows, err := r.queries.ListEntries(ctx)
	if err != nil {
		return nil, domain.ErrStorageFailure
	}

	var entries []domain.Entry
	for _, row := range rows {
		entries = append(entries, domain.Entry{
			WebsiteURL: row.WebsiteUrl,
			Username:	row.Username,
		})
	}
	return entries, nil
}

func (r *SQLiteRepository) UpdateEntry(ctx context.Context, websiteurl, username string, encryptedpass []byte) error {
	err := r.queries.UpdateEntry(ctx, db.UpdateEntryParams{
		EncryptedPassword:	[]byte(encryptedpass),
		WebsiteUrl: 		websiteurl, 
		Username: 			username,
	})
	if err != nil {
		return domain.ErrStorageFailure
	}
	return nil
}

func (r *SQLiteRepository) DeleteEntry(ctx context.Context, website_url, username string) error {
	err := r.queries.DeleteEntry(ctx, db.DeleteEntryParams{
		WebsiteUrl: website_url,
		Username:   username,
	})
	if err != nil {
		return domain.ErrStorageFailure
	}
	return nil
}

func (r *SQLiteRepository) SearchEntries(ctx context.Context, keyword string) ([]domain.Entry, error) {
	searchTerm := "%" + keyword + "%"
	rows, err := r.queries.SearchEntries(ctx, searchTerm)
	if err != nil {
		return nil, domain.ErrStorageFailure
	}

	var entries []domain.Entry
	for _, row := range rows {
		entries = append(entries, domain.Entry{
			WebsiteURL:  row.WebsiteUrl,
			Username: row.Username,
		})
	}
	return entries, nil
}

func (r *SQLiteRepository) ImportEntry(ctx context.Context, website, username string, encryptedPass []byte) error {
	return r.queries.ImportEntry(ctx, db.ImportEntryParams{
		WebsiteUrl:        website,
		Username:          username,
		EncryptedPassword: encryptedPass,
	})
}

func (r *SQLiteRepository) SetConfig(ctx context.Context, key string, value []byte) error {
	return r.queries.SetConfig(ctx, db.SetConfigParams{
		Key:	key,
		Value:	value,
	})
}

func (r *SQLiteRepository) GetConfig(ctx context.Context, key string) ([]byte, error) {
	return r.queries.GetConfig(ctx, key)
}