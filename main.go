package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

var (
	source      = flag.String("s", "./archives", "where is archives at?")
	destination = flag.String("d", "./decompressed", "decompress to where?")
	done        = make(chan struct{})
	sema        chan struct{}
)

func cancelled() bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}

func incomingFiles(dir string) <-chan *File {
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

func handle(src, des string) {
	select {
	case sema <- struct{}{}: // acquire token
	case <-done:
		fmt.Println("Cancelled.")
	}
	defer func() { <-sema }() // release token
	var n sync.WaitGroup
	for f := range incomingFiles(src) {
		n.Add(1)
		go func(f *File) {
			n.Done()
			start := time.Now()
			err := Unarchive(f.path, filepath.Join(des, f.name))
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Printf("%s, %s, %d bytes\n", f.name, time.Since(start), f.size)
		}(f)
	}
	n.Wait()
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Printf("\nTime consumed: %d\n", time.Since(start))
		select {}
	}()
	flag.Parse()
	go func() {
		os.Stdin.Read(make([]byte, 1)) // read a single byte.
		close(done)
		os.Exit(0)
	}()
	cpuUseNum := runtime.NumCPU() - 1
	runtime.GOMAXPROCS(cpuUseNum)
	sema = make(chan struct{}, cpuUseNum)
	// go handle("./test", "./unarchives")
	go handle(*source, *destination)
}
