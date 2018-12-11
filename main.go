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

func init() {
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
	direcory := os.Args[1]
	files, err := ioutil.ReadDir(direcory)
	check(err)

	// map of file size to the first file name
	fileSizeToCheck := make(map[int64]map[string][]byte)
	// map of the identical files according to thier md5 hash
	identicalFiles := make(map[string][]string)

	for i := 0; i < len(files); i++ {
		if file := files[i]; !file.IsDir() {
			size := file.Size()
			name := file.Name()
			checkedFile, inMap := fileSizeToCheck[size]
			if inMap {
				currentFileHash := getHashForFile(direcory, name)
				CHECKLOOP:
					for check := range checkedFile {
						if checkedFile[check] == nil {
							checkedFile[check] = getHashForFile(direcory, check)
						}
						if bytes.Equal(checkedFile[check], currentFileHash) {
							hexString := hex.EncodeToString(currentFileHash)
							identicalSlice, inMap := identicalFiles[hexString]
							if inMap {
								identicalFiles[hexString] = append(identicalSlice, name)
							} else {
								identicalFiles[hexString] = []string{check, name}
							}
							break CHECKLOOP
						}
					}
					checkedFile[name] = nil
			} else {
				fileSizeToCheck[size] = map[string][]byte{name: nil}
			}
		}
	}
	printResult(identicalFiles)
}

func printResult(identicalFiles map[string][]string ) {
	// Print indentical files
	for k, v := range identicalFiles {
		log.Printf("The following files are identicals with the hash %v", k)
		for _, name := range v {
			log.Printf("--> %v", name)
		}
	}
}