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

type File struct {
	Path    string
	Name    string
	Size    int64
	ModTime time.Time
}

var Done = make(chan struct{})

func cancelled() bool {
	select {
	case <-Done:
		return true
	default:
		return false
	}
}

func IncomingFiles(dir string) <-chan *File {
	ch := make(chan *File)
	go func() {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n",
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
			fmt.Printf("error walking the path %q: %v\n", dir, err)
		}
		close(ch)
	}()
	return ch
}

func Unarchive(src, des string) error {
	if cancelled() {
		return fmt.Errorf("Cancelled.")
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
