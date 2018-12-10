package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func check(e error) {
	if e != nil {
		log.Print(e)
		panic(e)
	}
}

func argumentCheck() {
	// Checking if only executable name passed to the program without an argument
	if len(os.Args) == 1 {
		log.Print("Missing argument for files direcory, Exiting...")
		os.Exit(1)
	}
}

func getHashForFile(folder string, file string) []byte {
	f, err := os.Open(filepath.Join(folder, file))
	defer f.Close()
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

	// map of file size to the first file name
	sizeToFirstFileName := make(map[int64]string)
	// map of the identical files according to thier md5 hash
	identicalFiles := make(map[string][]string)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		size := file.Size()
		// Check if map has the size key
		previousFileName, inMap := sizeToFirstFileName[size]
		if !inMap {
			// Add the first file with the size
			sizeToFirstFileName[size] = file.Name()
		} else {
			//check hash of the current file
			currentFileHash := getHashForFile(filePath, file.Name())
			// check hash of the already in map file
			previousFileHash := getHashForFile(filePath, previousFileName)

			if bytes.Equal(currentFileHash, previousFileHash) {
				hashString := hex.EncodeToString(currentFileHash)
				// Check if this is the first hash
				slice, inMap := identicalFiles[hashString]
				if !inMap {
					// Insert new hash to map
					identicalFiles[hashString] = []string{previousFileName, file.Name()}
				} else {
					// Add identical file to the map
					identicalFiles[hashString] = append(slice, file.Name())
				}
			}
		}
	}

	// Print indentical files
	for k, v := range identicalFiles {
		log.Printf("The following files are identicals with the hash %v", k)
		for _, name := range v {
			log.Printf("--> %v", name)
		}
	}
}
