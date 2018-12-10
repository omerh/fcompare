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

func StartHashWorkers(fileChan <-chan string) <-chan *HashResult {
	rc := make(chan *HashResult, runtime.NumCPU())

	// spawn runtime.NumCPU() hash workers
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(runtime.NumCPU())

		for i := 0; i < runtime.NumCPU(); i++ {
			go func() {
				for path := range fileChan {
					f, err := os.Open(path)

					if err != nil {
						log.Print(err)
						continue
					}

					defer f.Close()

					h := md5.New()
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

func ListFiles() <-chan string {
	rc := make(chan string)
	go func() {
		for _, dir := range os.Args[1:] {
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

func HashFiles() map[string][]string {
	rc := make(map[string][]string)

	for s := range StartHashWorkers(ListFiles()) {
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
		v := rc[s.Hash]
		v = append(v, s.Path)
		rc[s.Hash] = v
	}

	return rc
}

func main() {
	if len(os.Args) == 1 {
		log.Fatal("Missing argument for files direcory, Exiting...")
	}

	results := HashFiles()
	for hash := range results {
		// if the len of our resultsMap[p] is greater than 1, then we have multiple
		// files with the same hash.  this means the files should be identical so
		// we print that entry.
		if len(results[hash]) > 1 {
			fmt.Printf("%s %s\n", hash, results[hash])
		}
	}
}
