// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package embeddings

import (
	"sort"

	"gorm.io/gorm"
)

// SearchResult represents a search result with similarity score
type SearchResult struct {
	Slug       string
	Similarity float32
}

// VectorSearch provides vector similarity search functionality
// This is a pure-Go implementation using cosine similarity
// For better performance with large datasets, consider using sqlite-vec (requires CGO)
type VectorSearch struct {
	db      *gorm.DB
	service *Service
}

// NewVectorSearch creates a new vector search instance
func NewVectorSearch(db *gorm.DB, service *Service) *VectorSearch {
	return &VectorSearch{
		db:      db,
		service: service,
	}
}

// Search finds the most similar vectors to the query
// Returns results sorted by similarity (highest first)
func (v *VectorSearch) Search(query []float32, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	// Load all embeddings from database
	var embeddings []Embedding
	if err := v.db.Find(&embeddings).Error; err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return []SearchResult{}, nil
	}

	// Calculate similarity for each embedding
	results := make([]SearchResult, 0, len(embeddings))
	for _, emb := range embeddings {
		vector := BytesToVector(emb.Vector)
		similarity := CosineSimilarity(query, vector)

		results = append(results, SearchResult{
			Slug:       emb.Slug,
			Similarity: similarity,
		})
	}

	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// SearchWithThreshold finds vectors with similarity above the threshold
func (v *VectorSearch) SearchWithThreshold(query []float32, threshold float32, limit int) ([]SearchResult, error) {
	results, err := v.Search(query, limit*2) // Get more to filter
	if err != nil {
		return nil, err
	}

	// Filter by threshold
	filtered := make([]SearchResult, 0, len(results))
	for _, r := range results {
		if r.Similarity >= threshold {
			filtered = append(filtered, r)
		}
	}

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

// Store stores a vector for a slug
func (v *VectorSearch) Store(slug string, vector []float32) error {
	// This is handled by the Service.GetEmbedding method
	// This method is provided for API completeness
	emb := Embedding{
		Slug:         slug,
		ContentHash:  "", // Will be set by caller
		ModelName:    "",
		ModelVersion: "",
		Dimensions:   len(vector),
		Vector:       VectorToBytes(vector),
	}

	return v.db.Save(&emb).Error
}

// Delete removes a vector for a slug
func (v *VectorSearch) Delete(slug string) error {
	return v.db.Where("slug = ?", slug).Delete(&Embedding{}).Error
}

// Count returns the number of indexed vectors
func (v *VectorSearch) Count() (int64, error) {
	var count int64
	err := v.db.Model(&Embedding{}).Count(&count).Error
	return count, err
}

// SemanticSearch performs semantic search using the embedding service
// This is the main entry point for semantic search functionality
type SemanticSearch struct {
	service *Service
	search  *VectorSearch
}

// NewSemanticSearch creates a new semantic search instance
func NewSemanticSearch(service *Service, search *VectorSearch) *SemanticSearch {
	return &SemanticSearch{
		service: service,
		search:  search,
	}
}

// Search performs a semantic search for the query text
func (s *SemanticSearch) Search(query string, limit int) ([]SearchResult, error) {
	if !s.service.IsEnabled() {
		return nil, nil
	}

	// Generate embedding for the query
	queryVector, err := s.service.client.Embed(query)
	if err != nil {
		return nil, err
	}

	// Search for similar vectors
	return s.search.Search(queryVector, limit)
}

// SearchWithThreshold performs semantic search with a minimum similarity threshold
func (s *SemanticSearch) SearchWithThreshold(query string, threshold float32, limit int) ([]SearchResult, error) {
	if !s.service.IsEnabled() {
		return nil, nil
	}

	queryVector, err := s.service.client.Embed(query)
	if err != nil {
		return nil, err
	}

	return s.search.SearchWithThreshold(queryVector, threshold, limit)
}

// HybridSearch combines keyword and semantic search results
// Returns results that match both keyword and semantic criteria
func (s *SemanticSearch) HybridSearch(query string, keywordMatches []string, limit int) ([]SearchResult, error) {
	semanticResults, err := s.Search(query, limit*2)
	if err != nil {
		return nil, err
	}

	// Create a set of keyword matches for fast lookup
	keywordSet := make(map[string]bool)
	for _, slug := range keywordMatches {
		keywordSet[slug] = true
	}

	// Score results: boost semantic results that also have keyword matches
	boostedResults := make([]SearchResult, 0, len(semanticResults))
	for _, r := range semanticResults {
		if keywordSet[r.Slug] {
			// Boost similarity for keyword matches
			r.Similarity = r.Similarity * 1.2
			if r.Similarity > 1.0 {
				r.Similarity = 1.0
			}
		}
		boostedResults = append(boostedResults, r)
	}

	// Re-sort after boosting
	sort.Slice(boostedResults, func(i, j int) bool {
		return boostedResults[i].Similarity > boostedResults[j].Similarity
	})

	if len(boostedResults) > limit {
		boostedResults = boostedResults[:limit]
	}

	return boostedResults, nil
}
