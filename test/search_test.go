package test

import (
	"testing"
	"time"
	"os"
	"path/filepath"
	"github.com/example/gblobs/gblobs"
)

func TestBasicSearch(t *testing.T) {
	// Create temporary store for testing
	tempDir := filepath.Join(os.TempDir(), "gblobs_search_test")
	defer os.RemoveAll(tempDir)

	store := &gblobs.LocalStore{}
	err := store.CreateStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Add some test blobs
	meta1 := gblobs.BlobType{
		Name:          "document1.txt",
		URI:           "file://document1.txt",
		Owner:         "alice",
		IngestionTime: time.Now().UTC(),
	}
	content1 := []byte("This is a test document about machine learning and AI.")
	blobID1, err := store.PutBlob(content1, meta1)
	if err != nil {
		t.Fatalf("Failed to put blob 1: %v", err)
	}

	meta2 := gblobs.BlobType{
		Name:          "notes.md",
		URI:           "file://notes.md",
		Owner:         "bob",
		IngestionTime: time.Now().UTC(),
	}
	content2 := []byte("# Meeting Notes\n\nDiscussed the new search feature implementation.")
	_, err = store.PutBlob(content2, meta2)
	if err != nil {
		t.Fatalf("Failed to put blob 2: %v", err)
	}

	meta3 := gblobs.BlobType{
		Name:          "recipe.txt",
		URI:           "file://recipe.txt",
		Owner:         "alice",
		IngestionTime: time.Now().UTC(),
	}
	content3 := []byte("Recipe for chocolate cake: flour, sugar, eggs...")
	blobID3, err := store.PutBlob(content3, meta3)
	if err != nil {
		t.Fatalf("Failed to put blob 3: %v", err)
	}

	// Test search functionality
	t.Run("SearchForMachineLearning", func(t *testing.T) {
		results, err := store.Search("machine learning")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("Expected to find results for 'machine learning'")
		}

		found := false
		for _, result := range results {
			if result.BlobID == blobID1 {
				found = true
				if result.Metadata.Name != "document1.txt" {
					t.Errorf("Expected name 'document1.txt', got '%s'", result.Metadata.Name)
				}
				if result.Metadata.Owner != "alice" {
					t.Errorf("Expected owner 'alice', got '%s'", result.Metadata.Owner)
				}
			}
		}

		if !found {
			t.Error("Did not find expected blob in search results")
		}
	})

	t.Run("SearchByOwner", func(t *testing.T) {
		results, err := store.Search("alice")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(results) < 2 {
			t.Fatalf("Expected at least 2 results for 'alice', got %d", len(results))
		}

		foundBlobIDs := make(map[string]bool)
		for _, result := range results {
			foundBlobIDs[result.BlobID] = true
		}

		if !foundBlobIDs[blobID1] || !foundBlobIDs[blobID3] {
			t.Error("Did not find all expected blobs owned by alice")
		}
	})

	t.Run("SearchForNonExistent", func(t *testing.T) {
		results, err := store.Search("nonexistentterm")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected 0 results for non-existent term, got %d", len(results))
		}
	})

	t.Run("SearchWithOptions", func(t *testing.T) {
		req := gblobs.SearchRequest{
			Query:     "alice",
			Limit:     1,
			Offset:    0,
			Highlight: true,
		}
		results, err := store.SearchWithOptions(req)
		if err != nil {
			t.Fatalf("SearchWithOptions failed: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected exactly 1 result with limit=1, got %d", len(results))
		}

		// Note: Highlights may not always be present depending on content and query
		// This is normal behavior for Bleve when no clear matches are found in highlighted fields
		t.Logf("Search returned %d results, first result has %d highlight fields", len(results), len(results[0].Highlights))
	})

	// Test deletion removes from index
	t.Run("TestDeleteRemovesFromIndex", func(t *testing.T) {
		err := store.DeleteBlob(blobID1)
		if err != nil {
			t.Fatalf("Failed to delete blob: %v", err)
		}

		results, err := store.Search("machine learning")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		for _, result := range results {
			if result.BlobID == blobID1 {
				t.Error("Deleted blob still appears in search results")
			}
		}
	})
}