package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestUnRar(t *testing.T) {
	tc := struct {
		filename  string
		extracted []string
		want      []string
	}{
		"./test/test.tar",
		[]string{},
		[]string{
			"test/tmp/Office comparison.pdf",
			"test/tmp/自述文件.txt",
			"test/tmp/license-zh-cn.rtf",
			"test/tmp/hunspell-license-zh-cn.txt",
			"test/tmp/biblio.dbf",
			"test/tmp/tmw.dbf",
			"test/tmp/tmsmart.dat",
		},
	}
	f := &File{path: tc.filename}
	err := unRar(f)
	if err != nil {
		log.Fatal(err)
	}
	tc.extracted = getlist("./test/tmp")
	if !sliceCmp(tc.want, tc.extracted) {
		t.Errorf("\nwant: %v\ngot: %v\n", tc.want, tc.extracted)
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
