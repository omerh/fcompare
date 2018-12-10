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
	if len(os.Args[1:]) == 0 {
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

	sizeToFirstFileName := make(map[int64]string)
	identicalFiles := make(map[string][]string)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		size := file.Size()
		previousFileName, inMap := sizeToFirstFileName[size]
		if !inMap {
			sizeToFirstFileName[size] = file.Name()
		} else {
			currentFileHash := getHashForFile(filePath, file.Name())
			previousFileHash := getHashForFile(filePath, previousFileName)

			if bytes.Equal(currentFileHash, previousFileHash) {
				hashString := hex.EncodeToString(currentFileHash)
				slice, inMap := identicalFiles[hashString]
				if !inMap {
					identicalFiles[hashString] = []string{previousFileName, file.Name()}
				} else {
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
