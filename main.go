package main

import (
	"crypto/md5"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
)

type fileInfo struct {
	name string
	size int64
}

type filesInfo struct {
	fileInfo []fileInfo
}

type HashResult struct {
	Hash []byte
	Path string
}

func startHashWorkers(fileChan <-chan string) <-chan *HashResult {
	rc := make(chan *HashResult)
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
					rc <- &HashResult{Hash: h.Sum(nil), Path: path}
				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(rc)
	}()
	return rc
}

func main() {
	log.Print("starting app")

	if len(os.Args[1:]) == 0 {
		log.Fatal("Missing argument for files direcory, Exiting...")
	}

	listFiles := func() <-chan string {
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

	resultsMap := make(map[string][]byte)
	for s := range startHashWorkers(listFiles()) {
		resultsMap[s.Path] = s.Hash
	}

	for p := range resultsMap {
		log.Printf("%#v", resultsMap[p])
	}
}
