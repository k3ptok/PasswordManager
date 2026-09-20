package main

import (
	"database/sql"
	"log/slog"
	"embed"
	"os"

	"PasswordManager/internal/cmd"
	"PasswordManager/internal/db"
	"PasswordManager/internal/store"
	"PasswordManager/logger"
	"PasswordManager/config"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

//go:embed sql/schema/*.sql
var embedMigrations embed.FS

func main() {
	if err := run(); err != nil {
		slog.Error("Fatal Application Error", "error", "err")
		os.Exit(1)
	}
}

func run() error {
	// load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return err
	}
	logger.InitLogger(cfg.LogPath)
	slog.Info("Starting password manager...")
	dbDSN := cfg.DBPath + "?_foreign_keys=on"
	conn, err := sql.Open("sqlite3", dbDSN)
	if err != nil {
		slog.Error("Failed to open db connection", "error", err)
		return err
	}
	defer conn.Close()
	if err := os.Chmod(cfg.DBPath, 0600); err != nil && !os.IsNotExist(err) {
    	slog.Warn("Could not enforce strict database permissions", "error", err)
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		slog.Error("Failed to set goose dialect", "error", err)
		return err
	}
	if err := goose.Up(conn, "sql/schema"); err != nil {
		slog.Error("Database migration failed", "error", err)
		return err
	}
	slog.Info("Database migration successful")


	sqlcQueries := db.New(conn)
	repo := store.NewSQLiteRepo(sqlcQueries)

	appCLI := cmd.NewCLI(repo, cfg)
	
	return appCLI.Execute()

}