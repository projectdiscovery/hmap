package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/projectdiscovery/hmap/filekv"
	fileutil "github.com/projectdiscovery/utils/file"
)

func main() {
	// create 4 lists
	// two on disk
	list1, list2, list3 := "list1.txt", "list2.txt", "list3.txt"
	fList1, _ := os.Create(list1)
	for i := 0; i < 100000; i++ {
		_, _ = fList1.WriteString(fmt.Sprintf("%d\n", i))
	}
	fList1.Close()

	fList2, _ := os.Create(list2)
	// 1000 items overlaps
	for i := 90000; i < 200000; i++ {
		_, _ = fList2.WriteString(fmt.Sprintf("%d\n", i))
	}
	fList2.Close()

	// third list is still on disk but will be used as io.reader
	fList3, _ := os.Create(list3)
	// 1000 items overlaps
	for i := 190000; i < 300000; i++ {
		_, _ = fList3.WriteString(fmt.Sprintf("%d\n", i))
	}
	fList3.Close()

	// 4th list will be a list of new-line separated numbers
	var list4 strings.Builder
	for i := 290000; i < 400000; i++ {
		list4.WriteString(fmt.Sprintf("%d\n", i))
	}

	opts := filekv.DefaultOptions
	opts.Cleanup = true
	opts.Compress = true
	opts.Dedupe = filekv.MemoryLRU
	opts.SkipEmpty = true
	opts.Path = "fkv"
	fkv, err := filekv.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer fkv.Close()

	// opens various reader types
	flist3, _ := os.Open(list3)

	// add the various lists
	if _, err := fkv.Merge(list1, list2, flist3, strings.Split(list4.String(), filekv.NewLine)); err != nil {
		log.Fatal(err)
	}

	if err := fkv.Process(); err != nil {
		log.Fatal(err)
	}

	// scan the content
	readCount := 0
	err = fkv.Scan(func(b1, b2 []byte) error {
		readCount++
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fileutil.RemoveAll(list1, list2, list3)
	if readCount != 400000 {
		log.Fatalf("Expected 400000 but got %d\n", readCount)
	} else {
		log.Printf("Expected 400000 got %d\n", readCount)
	}
}
