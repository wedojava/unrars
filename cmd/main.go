package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/wedojava/unrars"
)

var (
	source      = flag.String("s", "./", "where is archives at?")
	destination = flag.String("d", "./_decompressed", "decompress to where?")
	max         = flag.Int("max", 500, "Coroutine upper limit!") // limit channel quantity
	sema        chan struct{}                                    // limit channel quantity
)

func handle(src, des string) {
	start := time.Now()
	select {
	case sema <- struct{}{}: // acquire token// limit channel quantity
	case <-unrars.Done:
		fmt.Println("[!] Cancelled.")
	}
	defer func() { <-sema }() // release token// limit channel quantity
	var n sync.WaitGroup
	for f := range unrars.IncomingFiles(src) {
		n.Add(1)
		go func(f *unrars.File) {
			defer n.Done()
			start := time.Now()
			// to avoid duplicating, this Unarchive will pour files to the subfolder
			// named with the archive file name
			// err := unrars.Unarchive(f.Path, filepath.Join(des, f.Name))

			// ignore duplicating, pour all files to the same folder: des
			err := unrars.Unarchive(f.Path, des)
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
	if cpuUseNum > 0 {
		runtime.GOMAXPROCS(cpuUseNum)
	}
	sema = make(chan struct{}, *max) // limit channel quantity
	// go handle("./test", "./unarchives")
	go handle(*source, *destination)
}
