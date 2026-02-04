// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package tools

import (
	"github.com/tejzpr/medha-mcp/internal/database"
	"github.com/tejzpr/medha-mcp/internal/git"
	"github.com/tejzpr/medha-mcp/internal/memory"
	"gorm.io/gorm"
)

// Query constants
const (
	querySlugEquals = "slug = ?"
)

// ToolContext holds shared dependencies for all tools
// In v2 architecture:
// - SystemDB: Global database for users, auth, repos (stays at ~/.medha/db/)
// - UserDB: Per-user database in .medha/medha.db inside git repo
// - DB: Kept for backward compatibility, points to SystemDB
type ToolContext struct {
	DB       *gorm.DB // Backward compatibility - points to SystemDB
	SystemDB *gorm.DB // Global database for users, auth, repos
	UserDB   *gorm.DB // Per-user database in .medha/medha.db
	RepoPath string
	DBMgr    *database.Manager // Database manager for handling connections
}

// NewToolContext creates a new tool context (v1 backward compatibility)
func NewToolContext(db *gorm.DB, repoPath string) *ToolContext {
	return &ToolContext{
		DB:       db,
		SystemDB: db,
		RepoPath: repoPath,
	}
}

// NewToolContextV2 creates a new tool context with separate system and user databases
func NewToolContextV2(systemDB *gorm.DB, userDB *gorm.DB, repoPath string) *ToolContext {
	return &ToolContext{
		DB:       systemDB, // Backward compatibility
		SystemDB: systemDB,
		UserDB:   userDB,
		RepoPath: repoPath,
	}
}

// NewToolContextWithManager creates a new tool context using the database manager
func NewToolContextWithManager(mgr *database.Manager, repoPath string) (*ToolContext, error) {
	userDB, err := mgr.GetUserDB(repoPath)
	if err != nil {
		return nil, err
	}

	return &ToolContext{
		DB:       mgr.SystemDB(), // Backward compatibility
		SystemDB: mgr.SystemDB(),
		UserDB:   userDB,
		RepoPath: repoPath,
		DBMgr:    mgr,
	}, nil
}

// GetRepository opens the git repository for operations
func (tc *ToolContext) GetRepository() (*git.Repository, error) {
	return git.OpenRepository(tc.RepoPath)
}

// GetMemoryBySlug retrieves a memory from the database by slug
// In v2, this queries the per-user database (UserDB)
// Falls back to v1 behavior (global DB) if UserDB is not set
func (tc *ToolContext) GetMemoryBySlug(slug string) (*database.MedhaMemory, error) {
	var mem database.MedhaMemory
	db := tc.DB
	if tc.UserDB != nil {
		// v2: Query from per-user database
		// Note: This returns MedhaMemory for backward compatibility
		// Internally, UserMemory is used in the per-user DB
		var userMem database.UserMemory
		err := tc.UserDB.Where(querySlugEquals, slug).First(&userMem).Error
		if err != nil {
			return nil, err
		}
		// Convert to MedhaMemory for backward compatibility
		mem = database.MedhaMemory{
			ID:             userMem.ID,
			Slug:           userMem.Slug,
			Title:          userMem.Title,
			FilePath:       userMem.FilePath,
			CreatedAt:      userMem.CreatedAt,
			UpdatedAt:      userMem.UpdatedAt,
			DeletedAt:      userMem.DeletedAt,
			SupersededBy:   userMem.SupersededBy,
			LastAccessedAt: userMem.LastAccessedAt,
			AccessCount:    userMem.AccessCount,
		}
		return &mem, nil
	}
	// v1 fallback: Query from global database
	err := db.Where(querySlugEquals, slug).First(&mem).Error
	return &mem, err
}

// GetUserMemoryBySlug retrieves a UserMemory from the per-user database by slug
// This is the v2 native method
func (tc *ToolContext) GetUserMemoryBySlug(slug string) (*database.UserMemory, error) {
	if tc.UserDB == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var mem database.UserMemory
	err := tc.UserDB.Where(querySlugEquals, slug).First(&mem).Error
	return &mem, err
}

// GetOrganizer returns a memory organizer for the repository
func (tc *ToolContext) GetOrganizer() *memory.Organizer {
	return memory.NewOrganizer(tc.RepoPath)
}

// HasUserDB returns true if the tool context has a per-user database
func (tc *ToolContext) HasUserDB() bool {
	return tc.UserDB != nil
}

// CloseUserDB closes the per-user database connection
// This should be called before git sync operations
func (tc *ToolContext) CloseUserDB() error {
	if tc.DBMgr != nil {
		return tc.DBMgr.CloseUserDB(tc.RepoPath)
	}
	if tc.UserDB != nil {
		sqlDB, err := tc.UserDB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// ReopenUserDB reopens the per-user database connection
// This should be called after git sync operations
func (tc *ToolContext) ReopenUserDB() error {
	if tc.DBMgr != nil {
		userDB, err := tc.DBMgr.ReopenUserDB(tc.RepoPath)
		if err != nil {
			return err
		}
		tc.UserDB = userDB
	}
	return nil
}
