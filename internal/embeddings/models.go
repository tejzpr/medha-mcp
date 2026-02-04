// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package embeddings

import (
	"time"

	"gorm.io/gorm"
)

// Embedding represents a stored embedding vector for a memory
type Embedding struct {
	Slug         string    `gorm:"primaryKey" json:"slug"`
	ContentHash  string    `gorm:"not null" json:"content_hash"`
	ModelName    string    `gorm:"not null" json:"model_name"`
	ModelVersion string    `gorm:"not null" json:"model_version"`
	Dimensions   int       `gorm:"not null" json:"dimensions"`
	Vector       []byte    `gorm:"type:blob;not null" json:"-"` // Stored as binary
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
}

// TableName specifies the table name for Embedding
func (Embedding) TableName() string {
	return "embeddings"
}

// MigrateEmbeddings runs migrations for the embeddings table
func MigrateEmbeddings(db *gorm.DB) error {
	return db.AutoMigrate(&Embedding{})
}

// CreateEmbeddingIndexes creates indexes for the embeddings table
func CreateEmbeddingIndexes(db *gorm.DB) error {
	indexes := []struct {
		name    string
		columns string
	}{
		{"idx_embeddings_content_hash", "slug, content_hash"},
		{"idx_embeddings_model", "model_name, model_version"},
	}

	for _, idx := range indexes {
		hasIndex := db.Migrator().HasIndex("embeddings", idx.name)
		if !hasIndex {
			sql := "CREATE INDEX IF NOT EXISTS " + idx.name + " ON embeddings (" + idx.columns + ")"
			if err := db.Exec(sql).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
