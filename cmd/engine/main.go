package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Pix3l-Guy21/search-engine/internal/crawler"
	"github.com/Pix3l-Guy21/search-engine/internal/indexer"
	"github.com/Pix3l-Guy21/search-engine/internal/parser"
	"github.com/Pix3l-Guy21/search-engine/internal/store"
	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Setup DB
	db, err := store.NewBadgerStore("./search_data")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	idx := indexer.NewIndexer(db)
	ctx := context.Background()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Println(".env Modified")
					runIndexing(ctx, db, idx)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				if err != nil {
					log.Println("Error in watcher: ", err)
				}
			}
		}
	}()

	err = watcher.Add(".env")
	if err != nil {
		log.Fatal(err)
	}

	var query string
	fmt.Println("\n--- Search Mode ---")
	fmt.Print("Type the word to search: ")
	fmt.Scan(&query)
	fmt.Println("Searching....")
	results, _ := idx.Search(ctx, query)

	if len(results) > 0 {
		for i, id := range results {
			name, _ := db.GetDocumentName(id)
			fmt.Printf("Result %d: %s (ID: %s)\n", i, name, id)
		}
	} else {
		fmt.Printf("No results found for '%s'\n", query)
	}
}

// Move your worker logic into this helper function
func runIndexing(ctx context.Context, db *store.BadgerStore, idx *indexer.Indexer) {
	done := make(chan struct{})
	envs, err := godotenv.Read(".env")
	if err != nil {
		log.Printf("Error at dotenv: %s", err)
	} else {
		log.Printf("Env: %s", envs["ROOT"])
	}
	root := envs["ROOT"]
	paths := crawler.Scan(done, root)
	docsChan := parser.Parser(done, paths)

	fmt.Println("Indexing started...")
	start := time.Now()

	var wg sync.WaitGroup
	const numWorkers = 1 // Keep at 1 for stability on large PDFs
	var count uint64

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for doc := range docsChan {
				log.Printf("[Worker %d] STARTING: %s", workerID, doc.Title)

				// CRITICAL: Added a check to skip the file that causes crashes
				if doc.Title == "jp547043.pdf" {
					log.Printf("[Worker %d] SKIPPING FAULTY FILE: %s", workerID, doc.Title)
					continue
				}

				extracted, err := parser.ExtractTextFromPdf(doc.Path)
				if err != nil {
					continue
				}
				doc.Tokens = parser.Tokenize(extracted)
				idx.IndexDocument(ctx, doc)

				atomic.AddUint64(&count, 1)
				log.Printf("[Worker %d] FINISHED: %s", workerID, doc.Title)
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("Finished! Indexed %d documents in %v\n", count, time.Since(start))
}
