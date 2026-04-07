package crawler

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

func Scan(done <-chan struct{}, root string) <-chan string {
	paths := make(chan string)
	extension := ".pdf"

	go func() {
		defer close(paths)
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(path) == extension {
				select {
				case paths <- path:
				case <-done:
					return fmt.Errorf("walk cancelled")
				}
			}
			return nil
		})
	}()

	return paths
}
