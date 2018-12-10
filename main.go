package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
)

type HashResult struct {
	Hash string
	Path string
}

func StartHashWorkers(fileChan <-chan string, numWorkers int) <-chan *HashResult {
	rc := make(chan *HashResult, numWorkers)

	// spawn numWorkers hash workers
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(numWorkers)

		for i := 0; i < numWorkers; i++ {
			go func() {
				h := md5.New()
				for path := range fileChan {
					f, err := os.Open(path)

					if err != nil {
						log.Print(err)
						continue
					}

					defer f.Close()

					h.Reset()
					if _, err := io.Copy(h, f); err != nil {
						log.Print(err)
						continue
					}
					// turn hash into a string. we do this to give us a text representation of
					// the hash AND to give us a comparable value to use as a map index.
					hashString := hex.EncodeToString(h.Sum(nil))
					rc <- &HashResult{Hash: hashString, Path: path}
				}
				wg.Done()
			}()
		}
		wg.Wait()

		// signal to the caller that we have no more results to send.
		close(rc)
	}()
	return rc
}

func ListFiles(dirs []string) <-chan string {
	rc := make(chan string)
	go func() {
		for _, dir := range dirs {
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				log.Printf("%v", err)
				continue
			}
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				rc <- path.Join(dir, file.Name())
			}
		}
		close(rc)
	}()
	return rc
}

func HashFiles(hr <-chan *HashResult) map[string][]string {
	rc := make(map[string][]string)

	for s := range hr {
		// this is subtle.  we attempt to fetch a slice of paths from []resultsMap.
		// if no entry exists, then we get back the zero value of a slice of
		// strings which is a nil slice.  otherwise we get back the stored slice of
		// strings.
		//
		// either way, we unconditionally append to that slice.  in the case that v
		// is a nil slice, this creates a new, single value.  if v is not nil, we
		// add another path to the existing slice of paths with the same hash.
		//
		// finally we unconditionally set the hash entry for the hash value to the
		// new slice of paths, thus adding our new path to the map at that hash's
		// map entry.
		//
		// The code below is equivalent to the following:
		//
		// v := rc[s.Hash]
		// v = append(v, s.Path)
		// rc[s.Hash] = v
		//
		rc[s.Hash] = append(rc[s.Hash], s.Path)
	}

	return rc
}

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("\n%[1]s - print identical files in specified directories.\n\n\tusage: %[1]s path1 ... pathN\n\n", os.Args[0])
		os.Exit(1)
	}

	for hash, paths := range HashFiles(StartHashWorkers(ListFiles(os.Args[1:]), runtime.NumCPU())) {
		if len(paths) == 1 {
			continue
		}
		fmt.Printf("%s %s\n", hash, paths)
	}
}
