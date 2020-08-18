package unrars

import (
	"compress/bzip2"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/klauspost/readahead"
)

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

// 38.51655895s
func bz2Decompress2(src, des string) error {
	start := time.Now()
	if err := os.MkdirAll(des, 0755); err != nil {
		return err
	}
	filename := filepath.Join(des, getFilename(src))
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	// save to file
	bzReader := bzip2.NewReader(f)
	buf := make([]byte, 1024*1000)
	w, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
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

// 42.738363312s
func bz2Decompress3(src, des string) error {
	start := time.Now()
	if err := os.MkdirAll(des, 0755); err != nil {
		return err
	}
	filename := filepath.Join(des, getFilename(src))
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	bzReader := bzip2.NewReader(f)
	w, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(2)

	rpipe, wpipe := io.Pipe()

	go func() {
		// write everything into the pipe. Decompression happens in this goroutine.
		io.Copy(wpipe, bzReader)
		wpipe.Close()
		wg.Done()
	}()
	go func() {
		for {
			n, err := io.Copy(w, rpipe)
			if n == 0 || err != nil {
				break
			}
		}
		wg.Done()
	}()

	wg.Wait()
	fmt.Println(time.Since(start))
	return nil
}

// 33.838901454s
func bz2Decompress(src, des string) error {
	// start := time.Now()
	if err := os.MkdirAll(des, 0755); err != nil {
		return err
	}

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	filename := filepath.Join(des, getFilename(src))
	w, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer w.Close()

	bzReader := bzip2.NewReader(f)
	ra := readahead.NewReader(bzReader)
	defer ra.Close()
	io.Copy(w, ra)
	// fmt.Println(time.Since(start))
	return nil
}

// study for error treating elegant
// 33.51906544s
func bz2Decompress1(src, des string) error {
	start := time.Now()
	var bzReader io.Reader
	var f, w *os.File
	var err error
	filename := filepath.Join(des, getFilename(src))
	err = func() error { // file obj prepare
		if err = os.MkdirAll(des, 0755); err != nil {
			return err
		}
		f, err = os.Open(src)
		if err != nil {
			return err
		}
		w, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		return err
	}
	defer f.Close()
	defer w.Close()

	bzReader = bzip2.NewReader(f)
	ra := readahead.NewReader(bzReader) // use readahead to optimize
	defer ra.Close()

	io.Copy(w, ra)
	fmt.Println(time.Since(start))
	return nil
}
