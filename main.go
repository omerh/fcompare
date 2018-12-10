package main

import (
	"bytes"
	"crypto/md5"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type fileInfo struct {
	name string
	size int64
}

type filesInfo struct {
	fileInfo []fileInfo
}

func check(e error) {
	if e != nil {
		log.Print(e)
		panic(e)
	}
}

func argumentCheck() {
	if len(os.Args[1:]) == 0 {
		log.Print("Missing argument for files direcory, Exiting...")
		os.Exit(1)
	}
}

func getHashForFile(folder string, file string) <-chan []byte {

	rc := make(chan []byte)

	go func() {
		f, err := os.Open(filepath.Join(folder, file))
		check(err)
		h := md5.New()
		if _, err := io.Copy(h, f); err != nil {
			check(err)
		}

		rc <- h.Sum(nil)
	}()

	return rc
}

func main() {
	log.Print("starting app")
	argumentCheck()
	filePath := os.Args[1]
	files, err := ioutil.ReadDir(filePath)
	check(err)

	filesInformation := []*fileInfo{}

	for _, file := range files {
		if !file.IsDir() {
			fi := fileInfo{
				name: file.Name(),
				size: file.Size(),
			}
			// log.Printf("Adding to slice %v", fileInfo{file.Name(), file.Size(), h.Sum(nil)})
			filesInformation = append(filesInformation, &fi)
		}
	}
	// Creating files list for comparison in order to delete original file information list
	compareFilesSlice := filesInformation

	for _, s := range filesInformation {
		for _, d := range compareFilesSlice {
			if s.name == d.name || s.size != d.size {
				continue
			}

			sH := getHashForFile(filePath, s.name)
			dH := getHashForFile(filePath, d.name)
			// log.Printf("Comprating checksum of %v with %x to %v with %x", s.name, s.checksum, d.name, d.checksum)
			if bytes.Equal(<-sH, <-dH) {
				log.Printf("Files %v and %v are identical with size %v and hash of %x", s.name, d.name, s.size, sH)
			}
		}
		// Removing from original slice the item that was compared
	}
}
