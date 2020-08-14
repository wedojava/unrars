package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestUnarchive(t *testing.T) {
	tcs := []struct {
		filename string
		want     []string
	}{
		{"./test/test.7z", []string{"test/tmp/7z"}},
		{"./test/test.tar", []string{"test/tmp/tar"}},
		{"./test/test.rar", []string{"test/tmp/rar"}},
		{"./test/test.tar.bz2", []string{"test/tmp/tarbz2"}},
		{"./test/test.tar.gz", []string{"test/tmp/targz"}},
		// {"./test/test.img.bz2", []string{"test/tmp/test.img"}},
	}
	// des, err := ioutil.TempDir("", "archiver_test")
	des := "test/tmp"
	for _, tc := range tcs {
		f := &File{path: tc.filename}
		err := Unarchive(f.path, des)
		if err != nil {
			log.Fatal(err)
		}
		got := getlist(des)
		if !sliceCmp(tc.want, got) {
			t.Errorf("\nwant: %v\ngot: %v\n", tc.want, got)
		}
		os.RemoveAll(des)
	}
}

func getlist(dir string) []string {
	res := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Size() != 0 {
			res = append(res, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func sliceCmp(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	i := 0
	for _, v1 := range s1 {
		for _, v2 := range s2 {
			if v1 == v2 {
				i++
			}
		}
	}
	if i == len(s1) {
		return true
	}
	fmt.Println(i)
	return false
}
