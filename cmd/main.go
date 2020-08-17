package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/wedojava/unrars"
)

var (
	source      = flag.String("s", "./", "where is archives at?")
	destination = flag.String("d", "./_decompressed", "decompress to where?")
	max         = flag.Int("max", 20, "Coroutine upper limit!")
	sema        chan struct{}
)

func handle(src, des string) {
	start := time.Now()
	select {
	case sema <- struct{}{}: // acquire token
	case <-unrars.Done:
		fmt.Println("[!] Cancelled.")
	}
	defer func() { <-sema }() // release token
	var n sync.WaitGroup
	for f := range unrars.IncomingFiles(src) {
		n.Add(1)
		go func(f *unrars.File) {
			defer n.Done()
			start := time.Now()
			err := unrars.Unarchive(f.Path, filepath.Join(des, f.Name))
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Printf("[+] %s, %s, %d bytes\n", f.Name, time.Since(start), f.Size)
		}(f)
	}
	n.Wait()
	fmt.Printf("\n[+] Time consumed: %v\n", time.Since(start))
	fmt.Printf("\n[+] Done.\n")
}

func main() {
	defer func() {
		select {}
	}()
	flag.Parse()
	go func() {
		os.Stdin.Read(make([]byte, 1)) // read a single byte.
		close(unrars.Done)
		os.Exit(0)
	}()
	cpuUseNum := runtime.NumCPU() - 1
	runtime.GOMAXPROCS(cpuUseNum)
	// TODO: if cpu cannot be used 100% on working, rm this assignment.
	*max = cpuUseNum
	sema = make(chan struct{}, *max)
	// go unrars.Handle("./test", "./unarchives")
	go handle(*source, *destination)
}
