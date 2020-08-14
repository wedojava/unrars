package main

import (
	"bufio"
	"compress/bzip2"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/klauspost/readahead"
	"github.com/mholt/archiver"
	"github.com/wedojava/go-unarr"
)

var (
	source      = flag.String("s", "./", "where is archives at?")
	destination = flag.String("d", "./_decompressed", "decompress to where?")
	done        = make(chan struct{})
	sema        chan struct{}
)

type File struct {
	path    string
	name    string
	size    int64
	modTime time.Time
}

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

func bz2Decompress1(src, des string) error {
	// open bzip2 file
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	// mkdir des
	if err := os.MkdirAll(des, 0755); err != nil {
		return err
	}
	// create file for save
	filename := filepath.Join(des, getFilename(src))
	// save to file
	bzReader := bzip2.NewReader(f)
	w, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	ra := readahead.NewReader(bzReader)
	defer ra.Close()
	s := bufio.NewScanner(ra)
	// s := bufio.NewScanner(bzReader)
	start := time.Now()
	for s.Scan() {
		w.Write(s.Bytes())
	}
	fmt.Println(time.Since(start))
	return nil
}

func savefile(bzipfile io.Reader, w io.Writer) error {
	// create a reader, using the bzip2.reader we were passed
	d := bufio.NewReader(bzipfile)
	// create a scanner
	s := bufio.NewScanner(d)
	// scan the file! until Scan() returns "EOF", print out each line
	for s.Scan() {
		w.Write(s.Bytes())
	}
	// we're done. return.
	return nil
}

func bz2Decompress2(src, des string) error {
	// open the file
	f, err := os.OpenFile(src, 0, 0)
	if err != nil {
		return nil
	}
	defer f.Close()
	// mkdir des
	if err := os.MkdirAll(des, 0755); err != nil {
		return err
	}
	// create file for save
	filename := filepath.Join(des, getFilename(src))
	// save to file
	w, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer w.Close()
	// create a reader
	br := bufio.NewReader(f)
	// create a bzip2.reader, using the reader we just created
	cr := bzip2.NewReader(br)
	// lets do this. parse the file.
	err = savefile(cr, w)
	return nil
}

// 37.496846188s
func bz2Decompress(src, des string) error {
	// open bzip2 file
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	// mkdir des
	if err := os.MkdirAll(des, 0755); err != nil {
		return err
	}
	// create file for save
	filename := filepath.Join(des, getFilename(src))
	// save to file
	bzReader := bzip2.NewReader(f)
	buf := make([]byte, 1024*1000)
	w, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	start := time.Now()
	for {
		n, err := bzReader.Read(buf)
		if n == 0 || err != nil {
			break
		}
		// write to file
		if _, err = w.Write(buf[:n]); err != nil {
			return err
		}
	}
	fmt.Println(time.Since(start))
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
	// go unrars.Handle("./test", "./unarchives")
	go handle(*source, *destination)
}
