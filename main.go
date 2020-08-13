package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gen2brain/go-unarr"
)

type File struct {
	path    string
	name    string
	size    int64
	modTime time.Time
}

func incomingFiles(dir string) <-chan *File {
	ch := make(chan *File)
	go func() {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if !info.IsDir() && info.Size() != int64(0) { // is file that not 0 bytes.
				ch <- &File{
					path:    path,
					name:    info.Name(),
					size:    info.Size(),
					modTime: info.ModTime(),
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("error walking the path %q: %v\n", dir, err)
		}
		close(ch)
	}()
	return ch
}

func unRar(f *File) error {
	a, err := unarr.NewArchive(f.path)
	// a, err := unarr.NewArchive("./test/test.7z")
	if err != nil {
		log.Printf("[-] unRar err: %v", err)
	}
	defer a.Close()
	lst, err := a.Extract("./test/tmp")
	if err != nil {
		log.Printf("[-] unRar extract err: %v", err)
	}
	fmt.Println(lst)
	return nil
}

func main() {

}
