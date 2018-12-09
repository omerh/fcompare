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

func getHashForFile(folder string, file string) []byte {
	f, err := os.Open(filepath.Join(folder, file))
	check(err)
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		check(err)
	}
	return h.Sum(nil)
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
			fi := new(fileInfo)
			fi.name = file.Name()
			fi.size = file.Size()

			// log.Printf("Adding to slice %v", fileInfo{file.Name(), file.Size(), h.Sum(nil)})
			filesInformation = append(filesInformation, fi)
		}
	}
	// Creating files list for comparison in order to delete original file information list
	compareFilesSlice := filesInformation

	for i := 0; i < len(filesInformation); i++ {
		var s = filesInformation[i]
		for _, d := range compareFilesSlice {
			if s.name != d.name {
				// log.Printf("Comparing file %v to %v", s.name, d.name)
				if s.size == d.size {
					sH := getHashForFile(filePath, s.name)
					dH := getHashForFile(filePath, d.name)
					// log.Printf("Comprating checksum of %v with %x to %v with %x", s.name, s.checksum, d.name, d.checksum)
					if bytes.Equal(sH, dH) {
						log.Printf("Files %v and %v are identical with size %v and hash of %x", s.name, d.name, s.size, sH)
					}
				}
			}
		}
		// Removing from original slice the item that was compared
		filesInformation = append(filesInformation[:i], filesInformation[i+1:]...)
	}
}
