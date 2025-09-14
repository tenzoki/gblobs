#!/bin/bash
set -e
echo "+ Search Demo: Full-text search capabilities"

STORE=tmp_search_demo_store
rm -rf "$STORE"

echo "=== Setting up test data ==="
echo "Creating sample documents..."

# Create sample documents
echo "This is a document about machine learning and artificial intelligence. It discusses neural networks, deep learning, and data science applications." > ml_doc.txt
echo "Meeting notes from the project review: We discussed the new search feature implementation using Bleve full-text search library." > meeting.txt
echo "Recipe for chocolate cake: Mix flour, sugar, eggs, and chocolate. Bake for 45 minutes at 350 degrees." > recipe.txt
echo "Shopping list: bread, milk, eggs, chocolate, apples, bananas" > shopping.txt

# Add documents to store
echo "Adding documents to gblobs store..."
DOC_ID=$(../cmd/gblobs/gblobs putfile --store "$STORE" --owner alice ml_doc.txt)
MEETING_ID=$(../cmd/gblobs/gblobs putfile --store "$STORE" --owner bob meeting.txt)
RECIPE_ID=$(../cmd/gblobs/gblobs putfile --store "$STORE" --owner alice recipe.txt)
SHOPPING_ID=$(../cmd/gblobs/gblobs putfile --store "$STORE" --owner charlie shopping.txt)

# Add some string content too
STRING_ID=$(../cmd/gblobs/gblobs putstring --store "$STORE" --owner bob --name "status.txt" "Project status: search feature is implemented and working well")

echo "Added 5 documents to the store"
echo ""

echo "=== Basic Search Tests ==="

echo "1. Search for 'machine learning':"
../cmd/gblobs/gblobs search --store "$STORE" "machine learning"
echo ""

echo "2. Search for 'chocolate':"
../cmd/gblobs/gblobs search --store "$STORE" "chocolate"
echo ""

echo "3. Search for 'search' (should find multiple results):"
../cmd/gblobs/gblobs search --store "$STORE" "search"
echo ""

echo "=== Owner-based Search ==="

echo "4. Search for content by 'alice':"
../cmd/gblobs/gblobs search --store "$STORE" "alice"
echo ""

echo "5. Search for content by 'bob':"
../cmd/gblobs/gblobs search --store "$STORE" "bob"
echo ""

echo "=== Advanced Search Options ==="

echo "6. Limited search results (limit=2):"
../cmd/gblobs/gblobs search --store "$STORE" --limit 2 "alice"
echo ""

echo "7. Search with pagination (offset=1):"
../cmd/gblobs/gblobs search --store "$STORE" --limit 1 --offset 1 "alice"
echo ""

echo "=== Search for Non-existent Terms ==="

echo "8. Search for non-existent term:"
../cmd/gblobs/gblobs search --store "$STORE" "nonexistent"
echo ""

echo "=== Search and Retrieve Workflow ==="

echo "9. Search for content and retrieve using blob ID:"
echo "   Search for 'machine learning':"
../cmd/gblobs/gblobs search --store "$STORE" "machine learning"

echo "   Retrieve the document using the blob ID from search results:"
echo "   Content:"
../cmd/gblobs/gblobs get --store "$STORE" "$DOC_ID"
echo ""

echo "=== Test Index Cleanup on Deletion ==="

echo "10. Delete a document and verify it's removed from search:"
echo "   Before deletion - search for 'neural networks':"
../cmd/gblobs/gblobs search --store "$STORE" "neural networks"

echo "   Deleting the machine learning document..."
../cmd/gblobs/gblobs delete --store "$STORE" "$DOC_ID"

echo "   After deletion - search for 'neural networks':"
../cmd/gblobs/gblobs search --store "$STORE" "neural networks"
echo ""

echo "=== Final Store Inspection ==="
echo "11. Inspect remaining blobs in store:"
../cmd/gblobs/gblobs inspect --store "$STORE"

echo "=== Store Statistics ==="
echo "12. Show store stats:"
../cmd/gblobs/gblobs stats --store "$STORE"

# Cleanup
rm -rf "$STORE" ml_doc.txt meeting.txt recipe.txt shopping.txt
echo ""
echo "Search demo completed successfully!"
echo "All test files and store cleaned up."