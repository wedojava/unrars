package unrars

import (
	"compress/bzip2"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/klauspost/readahead"
)

// 35.215696512s
func bz2Decompress2(src, des string) error {
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

// 36.360893309s
func bz2Decompress3(src, des string) error {
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

	start := time.Now()
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

// 32.65924852s
func bz2Decompress(src, des string) error {
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
	ra := readahead.NewReader(bzReader)
	defer ra.Close()
	start := time.Now()
	io.Copy(w, ra)
	fmt.Println(time.Since(start))
	return nil
}
