package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

// findMigrationsDir finds the migrations directory by walking up from the current directory
func findMigrationsDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Walk up the directory tree to find the project root
	dir := wd
	for {
		migrationsPath := filepath.Join(dir, "migrations")
		if info, err := os.Stat(migrationsPath); err == nil && info.IsDir() {
			return migrationsPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	// Fallback: try relative to current directory
	migrationsPath, err := filepath.Abs("migrations")
	if err != nil {
		return "", fmt.Errorf("failed to find migrations directory: %w", err)
	}
	return migrationsPath, nil
}

// EnsureDatabase creates the database if it doesn't exist
func EnsureDatabase(user, password, host, port, dbName string) error {
	// Connect to the default 'postgres' database to create the target database
	defaultDBURL := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", user, password, host, port)

	pool, err := pgxpool.New(context.Background(), defaultDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer pool.Close()

	// Check if database exists
	var exists bool
	err = pool.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		// Create the database
		// Escape the database name properly for CREATE DATABASE
		// Replace any double quotes in the name to prevent SQL injection
		safeDBName := strings.ReplaceAll(dbName, `"`, `""`)
		// Quote the identifier to handle special characters
		quotedDBName := fmt.Sprintf(`"%s"`, safeDBName)
		_, err = pool.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", quotedDBName))
		if err != nil {
			return fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
		fmt.Printf("✅ Created database: %s\n", dbName)
	}

	return nil
}

// RunMigrations applies all migrations
func RunMigrations(user, password, host, port, dbName string) error {
	// Ensure database exists before running migrations
	if err := EnsureDatabase(user, password, host, port, dbName); err != nil {
		return fmt.Errorf("failed to ensure database exists: %w", err)
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)

	// Find migrations directory
	migrationsPath, err := findMigrationsDir()
	if err != nil {
		return err
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Calculate relative path from current working directory to migrations
	relPath, err := filepath.Rel(wd, migrationsPath)
	if err != nil {
		// If relative path calculation fails, try absolute path
		migrationsPath, err = filepath.Abs(migrationsPath)
		if err != nil {
			return fmt.Errorf("failed to get migrations path: %w", err)
		}
		// Convert to forward slashes and format for file:// URL
		migrationsPath = filepath.ToSlash(migrationsPath)
		if len(migrationsPath) > 1 && migrationsPath[1] == ':' {
			// Windows: file:///C:/path
			fileURL := fmt.Sprintf("file:///%s", strings.TrimPrefix(migrationsPath, "/"))
			m, err := migrate.New(fileURL, dbURL)
			if err != nil {
				return fmt.Errorf("failed to create migrate instance: %w", err)
			}
			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				return fmt.Errorf("failed to run migrations: %w", err)
			}
			return nil
		}
		// Unix: file:///path
		fileURL := fmt.Sprintf("file://%s", migrationsPath)
		m, err := migrate.New(fileURL, dbURL)
		if err != nil {
			return fmt.Errorf("failed to create migrate instance: %w", err)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
		return nil
	}

	// Use relative path with forward slashes
	relPath = filepath.ToSlash(relPath)
	// Ensure it starts with ./ for relative paths
	if !strings.HasPrefix(relPath, ".") {
		relPath = "./" + relPath
	}
	fileURL := fmt.Sprintf("file://%s", relPath)

	m, err := migrate.New(fileURL, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
