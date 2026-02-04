// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Manager handles database connections for both system and per-user databases
// System DB: Global database for users, auth, and repo registry
// User DB: Per-user database stored in .medha/medha.db inside each git repo
type Manager struct {
	systemDB   *gorm.DB
	config     *Config
	userDBs    map[string]*gorm.DB
	userDBsMux sync.RWMutex
}

// NewManager creates a new database manager with a system database connection
func NewManager(cfg *Config) (*Manager, error) {
	systemDB, err := Connect(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system database: %w", err)
	}

	// Run system DB migrations
	if err := MigrateSystemDB(systemDB); err != nil {
		return nil, fmt.Errorf("failed to migrate system database: %w", err)
	}

	// Create system DB indexes
	if err := CreateSystemIndexes(systemDB); err != nil {
		return nil, fmt.Errorf("failed to create system indexes: %w", err)
	}

	return &Manager{
		systemDB: systemDB,
		config:   cfg,
		userDBs:  make(map[string]*gorm.DB),
	}, nil
}

// SystemDB returns the system database connection
func (m *Manager) SystemDB() *gorm.DB {
	return m.systemDB
}

// GetUserDB opens or returns an existing per-user database connection
// The per-user database is stored at {repoPath}/.medha/medha.db
func (m *Manager) GetUserDB(repoPath string) (*gorm.DB, error) {
	// Check cache first
	m.userDBsMux.RLock()
	if db, ok := m.userDBs[repoPath]; ok {
		m.userDBsMux.RUnlock()
		return db, nil
	}
	m.userDBsMux.RUnlock()

	// Open new connection
	m.userDBsMux.Lock()
	defer m.userDBsMux.Unlock()

	// Double-check after acquiring write lock
	if db, ok := m.userDBs[repoPath]; ok {
		return db, nil
	}

	db, err := OpenUserDB(repoPath)
	if err != nil {
		return nil, err
	}

	m.userDBs[repoPath] = db
	return db, nil
}

// OpenUserDB opens a per-user database at the specified repository path
// Creates the .medha directory and database if they don't exist
func OpenUserDB(repoPath string) (*gorm.DB, error) {
	dbPath := GetUserDBPath(repoPath)

	// Ensure .medha directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .medha directory: %w", err)
	}

	// Open SQLite database with git-friendly settings
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open user database: %w", err)
	}

	// Configure for git compatibility
	// Use DELETE journal mode instead of WAL to avoid multiple files
	// This ensures the database is a single file that can be tracked in git
	db.Exec("PRAGMA journal_mode = DELETE")
	db.Exec("PRAGMA synchronous = NORMAL")

	// Run migrations for user database
	if err := MigrateUserDB(db); err != nil {
		return nil, fmt.Errorf("failed to migrate user database: %w", err)
	}

	// Create indexes
	if err := CreateUserIndexes(db); err != nil {
		return nil, fmt.Errorf("failed to create user indexes: %w", err)
	}

	return db, nil
}

// GetUserDBPath returns the path to the per-user database
func GetUserDBPath(repoPath string) string {
	return filepath.Join(repoPath, ".medha", "medha.db")
}

// CloseUserDB closes a specific user database connection
func (m *Manager) CloseUserDB(repoPath string) error {
	m.userDBsMux.Lock()
	defer m.userDBsMux.Unlock()

	if db, ok := m.userDBs[repoPath]; ok {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		if err := sqlDB.Close(); err != nil {
			return err
		}
		delete(m.userDBs, repoPath)
	}
	return nil
}

// ReopenUserDB closes and reopens a user database connection
// Useful after git sync to ensure fresh state
func (m *Manager) ReopenUserDB(repoPath string) (*gorm.DB, error) {
	if err := m.CloseUserDB(repoPath); err != nil {
		return nil, err
	}
	return m.GetUserDB(repoPath)
}

// Close closes the system database and all user database connections
func (m *Manager) Close() error {
	// Close all user DBs
	m.userDBsMux.Lock()
	for path, db := range m.userDBs {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
		delete(m.userDBs, path)
	}
	m.userDBsMux.Unlock()

	// Close system DB
	return Close(m.systemDB)
}

// ConnectSystemDB connects to the system database based on configuration
// This is a standalone function for backward compatibility
func ConnectSystemDB(cfg *Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(cfg.LogLevel),
	}

	switch cfg.Type {
	case "sqlite":
		if err := ensureSQLiteDir(cfg.SQLitePath); err != nil {
			return nil, fmt.Errorf("failed to ensure sqlite directory: %w", err)
		}
		db, err = gorm.Open(sqlite.Open(cfg.SQLitePath), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
		}

	case "postgres":
		db, err = gorm.Open(postgres.Open(cfg.PostgresDSN), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to postgres: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	return db, nil
}

// UserDBExists checks if a per-user database exists at the specified path
func UserDBExists(repoPath string) bool {
	dbPath := GetUserDBPath(repoPath)
	_, err := os.Stat(dbPath)
	return err == nil
}

// GetJournalMode returns the current journal mode of a database
func GetJournalMode(db *gorm.DB) (string, error) {
	var mode string
	err := db.Raw("PRAGMA journal_mode").Scan(&mode).Error
	return mode, err
}
