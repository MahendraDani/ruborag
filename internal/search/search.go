package search

import (
	"log"
	"ruborag/internal/db"
	"ruborag/internal/similarity"
	"sort"
)

type SearchResult struct {
	SourceFile string
	ChunkIndex int
	Content    string
	Score      float32
}

func SearchQuery(queryVec []float32, embeddings []db.StoredEmbedding) []SearchResult {
	if len(embeddings) == 0 {
		log.Fatal("no embeddings found in database")
	}

	results := make([]SearchResult, 0, len(embeddings))

	for _, e := range embeddings {
		score := similarity.CosineSimilarity(queryVec, e.Vector)
		results = append(results, SearchResult{
			SourceFile: e.SourceFile,
			ChunkIndex: e.ChunkIndex,
			Content:    e.Content,
			Score:      score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}
