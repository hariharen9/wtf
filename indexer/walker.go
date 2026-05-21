package indexer

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"wtf/config"
)

type WalkItem struct {
	Path  string
	IsDir bool
}

// ParallelWalk walks starting from config roots, ignoring matching directories.
// It streams found paths to the outChan channel.
func ParallelWalk(cfg *config.Config, outChan chan<- WalkItem) error {
	dirChan := make(chan string, 200000)
	var activeTasks int64

	// Concurrency level tuned for modern SSDs and multi-core processors
	numWorkers := runtime.NumCPU() * 2
	if numWorkers < 8 {
		numWorkers = 8
	}

	var wg sync.WaitGroup

	// Push initial roots to the queue
	for _, root := range cfg.Roots {
		abs, err := filepath.Abs(root)
		if err != nil {
			abs = root
		}
		
		// Verify root exists
		info, err := os.Stat(abs)
		if err != nil {
			continue
		}

		if !cfg.ShouldIgnore(abs) {
			atomic.AddInt64(&activeTasks, 1)
			dirChan <- abs
			
			// We can also report the root itself
			outChan <- WalkItem{Path: abs, IsDir: true}
		} else {
			_ = info
		}
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dir := range dirChan {
				entries, err := os.ReadDir(dir)
				if err != nil {
					// In case of error (permission denied, etc.), decrement tasks and check for completion
					if atomic.AddInt64(&activeTasks, -1) == 0 {
						close(dirChan)
					}
					continue
				}

				for _, entry := range entries {
					name := entry.Name()
					fullPath := filepath.Join(dir, name)

					if cfg.ShouldIgnore(fullPath) {
						continue
					}

					isDir := entry.IsDir()
					
					// Stream the path
					outChan <- WalkItem{
						Path:  fullPath,
						IsDir: isDir,
					}

					if isDir {
						atomic.AddInt64(&activeTasks, 1)
						select {
						case dirChan <- fullPath:
						default:
							// If buffer is full, block and write.
							// Since it's buffered to 200k, blocking only happens on massive directory trees.
							dirChan <- fullPath
						}
					}
				}

				// Finished processing this directory
				if atomic.AddInt64(&activeTasks, -1) == 0 {
					close(dirChan)
				}
			}
		}()
	}

	wg.Wait()
	return nil
}
const (
	MaxWorkerQueue = 200000
)
