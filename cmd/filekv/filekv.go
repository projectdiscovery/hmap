package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/projectdiscovery/fileutil"
	"github.com/projectdiscovery/hmap/filekv"
)

func main() {
	// create 4 lists
	// two on disk
	list1, list2, list3 := "list1.txt", "list2.txt", "list3.txt"
	fList1, _ := os.Create(list1)
	for i := 0; i < 10000; i++ {
		_, _ = fList1.WriteString(fmt.Sprintf("%d\n", i))
	}
	fList1.Close()

	fList2, _ := os.Create(list2)
	// 1000 items overlaps
	for i := 9000; i < 20000; i++ {
		_, _ = fList2.WriteString(fmt.Sprintf("%d\n", i))
	}
	fList2.Close()

	// third list is still on disk but will be used as io.reader
	fList3, _ := os.Create(list3)
	// 1000 items overlaps
	for i := 19000; i < 30000; i++ {
		_, _ = fList3.WriteString(fmt.Sprintf("%d\n", i))
	}
	fList3.Close()

	// 4th list will be a list of new-line separated numbers
	var list4 strings.Builder
	for i := 29000; i < 40000; i++ {
		list4.WriteString(fmt.Sprintf("%d\n", i))
	}

	opts := filekv.DefaultOptions
	opts.Cleanup = true
	opts.Compress = true
	opts.Dedupe = true
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
	err = fkv.Scan(func(b1, b2 []byte) error {
		log.Print("read:", string(b1), string(b2))
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fileutil.RemoveAll(list1, list2, list3)
}
