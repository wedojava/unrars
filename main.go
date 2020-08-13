package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen2brain/go-unarr"
	"github.com/mholt/archiver/v3"
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

func Unarchive(src, des string) error {
	is7z := strings.HasSuffix(src, ".7z")
	if is7z {
		a, err := unarr.NewArchive(src)
		if err != nil {
			return err
		}
		_, err = a.Extract(des)
		if err != nil {
			return err
		}
	} else {
		err := archiver.Unarchive(src, des)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {

}
