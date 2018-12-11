package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

func getHashForFile(folder string, file string) []byte {
	f, err := os.Open(filepath.Join(folder, file))
	defer f.Close()
	if err != nil {
		panic(err)
	}
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		if err != nil {
			panic(err)
		}
	}
	return h.Sum(nil)
}

/// a thread safe wrapper for a map from file names to hash with the necessary data to compute the hash
type suspectSet struct {
	dir        string
	suspects   map[string][]byte
	mutex      *sync.Mutex
	identicals *identicalsSet
}

/// a thread safe wrapper for a map from hex strings to slice of files which hash into the relavant hex string
type identicalsSet struct {
	identicals map[string][]string
	mutex      *sync.Mutex
}

func (identicals *identicalsSet) addIdentical(hexString, sentinelFile, checkFile string) {
	identicals.mutex.Lock()
	defer identicals.mutex.Unlock()
	identicalSlice, inMap := identicals.identicals[hexString]
	if inMap {
		identicals.identicals[hexString] = append(identicalSlice, checkFile)
	} else {
		identicals.identicals[hexString] = []string{sentinelFile, checkFile}
	}
}

func (suspects *suspectSet) chcekSuspects(fileName string, wg *sync.WaitGroup) {
	suspects.mutex.Lock()
	defer func() {
		suspects.mutex.Unlock()
		wg.Done()
	}()

	currentHash := getHashForFile(suspects.dir, fileName)
	for suspect := range suspects.suspects {
		if suspects.suspects[suspect] == nil {
			suspects.suspects[suspect] = getHashForFile(suspects.dir, suspect)
		}
		if bytes.Equal(suspects.suspects[suspect], currentHash) {
			hexString := hex.EncodeToString(currentHash)
			suspects.identicals.addIdentical(hexString, suspect, fileName)
			break
		}
	}
	suspects.suspects[fileName] = currentHash
}

func findDuplicatesThreaded(dir string) map[string][]string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	sizeToSuspects := make(map[int64]*suspectSet)
	identicals := &identicalsSet{
		mutex:      &sync.Mutex{},
		identicals: make(map[string][]string),
	}
	wg := sync.WaitGroup{}

	for i := 0; i < len(files); i++ {
		if file := files[i]; !file.IsDir() {
			size := file.Size()
			name := file.Name()
			suspects, inMap := sizeToSuspects[size]
			if inMap {
				wg.Add(1)
				go suspects.chcekSuspects(name, &wg)
			} else {
				sizeToSuspects[size] = &suspectSet{
					dir:        dir,
					mutex:      &sync.Mutex{},
					suspects:   map[string][]byte{name: nil},
					identicals: identicals,
				}
			}
		}
	}
	wg.Wait()
	return identicals.identicals
}

func findDuplicates(dir string) map[string][]string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	sizeToSuspects := make(map[int64]map[string][]byte)
	identicals := make(map[string][]string)

	for i := 0; i < len(files); i++ {
		if file := files[i]; !file.IsDir() {
			size := file.Size()
			name := file.Name()
			suspectSet, inMap := sizeToSuspects[size]
			if inMap {
				currentHash := getHashForFile(dir, name)
			SUSPECTLOOP:
				for suspect := range suspectSet {
					if suspectSet[suspect] == nil {
						suspectSet[suspect] = getHashForFile(dir, suspect)
					}
					if bytes.Equal(suspectSet[suspect], currentHash) {
						hexString := hex.EncodeToString(currentHash)
						identicalSlice, inMap := identicals[hexString]
						if inMap {
							identicals[hexString] = append(identicalSlice, name)
						} else {
							identicals[hexString] = []string{suspect, name}
						}
						break SUSPECTLOOP
					}
				}
				suspectSet[name] = currentHash
			} else {
				sizeToSuspects[size] = map[string][]byte{name: nil}
			}
		}
	}
	return identicals
}

func cli() (bool, string) {
	var threaded = flag.Bool("t", false, "set to parralelize hash calculations")

	flag.Parse()

	if flag.NArg() > 1 {
		panic(fmt.Sprintf("received too many command line args: %s", flag.Args()))
	} else if flag.NArg() < 1 {
		panic("did not receive directory")
	}

	directory := flag.Arg(0)

	return *threaded, directory
}

func fcompare(threaded bool, dir string) {

	var identicals map[string][]string
	if threaded {
		identicals = findDuplicatesThreaded(dir)
	} else {
		identicals = findDuplicates(dir)
	}

	for k, v := range identicals {
		fmt.Println("The following files are identicals with the hash", k)
		for _, name := range v {
			fmt.Println("    ", name)
		}
	}
}

func main() {
	threaded, dir := cli()
	fcompare(threaded, dir)
}
