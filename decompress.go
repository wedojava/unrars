package unrars

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen2brain/go-unarr"
	"github.com/mholt/archiver"
)

// File is file contains many details we need to send to channel
type File struct {
	Path    string
	Name    string
	Size    int64
	ModTime time.Time
}

// Done is the channel for cancel
var Done = make(chan struct{})

func cancelled() bool {
	select {
	case <-Done:
		return true
	default:
		return false
	}
}

// IncomingFiles find out all files, by walk the dir and make Files,
// pull it to channel then return
func IncomingFiles(dir string) <-chan *File {
	ch := make(chan *File)
	go func() {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("[-] prevent panic by handling failure accessing a path %q: %v",
					path, err)
				return err
			}
			if !info.IsDir() && info.Size() != int64(0) { // is file that not 0 bytes.
				ch <- &File{
					Path:    path,
					Name:    info.Name(),
					Size:    info.Size(),
					ModTime: info.ModTime(),
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("[-] error walking the path %q: %v", dir, err)
		}
		close(ch)
	}()
	return ch
}

// Unarchive will treat file at src and unarchive it to des folder
func Unarchive(src, des string) error {
	if cancelled() {
		return fmt.Errorf("[-] cancelled")
	}
	is7z := strings.HasSuffix(src, ".7z")
	isbz2 := strings.HasSuffix(src, ".bz2") && !strings.HasSuffix(src, ".tar.bz2")
	switch {
	case is7z:
		a, err := unarr.NewArchive(src)
		if err != nil {
			return err
		}
		if _, err = a.Extract(des); err != nil {
			return err
		}
	case isbz2:
		if err := bz2Decompress(src, des); err != nil {
			return err
		}
	default:
		if err := archiver.Unarchive(src, des); err != nil {
			return err
		}
	}
	return nil
}

// getFilename extract filename without ".bz2" ext name from src
func getFilename(src string) string {
	rs := []rune(src)
	i := strings.LastIndex(src, "\\")
	if i == -1 {
		i = strings.LastIndex(src, "/")
	}
	res := string(rs[i+1:])
	res = strings.Split(res, ".bz2")[0]
	return res
}
