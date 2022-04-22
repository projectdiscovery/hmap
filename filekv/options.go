package filekv

import "math"

const Separator = ";;;"

type Options struct {
	Path     string
	Dedupe   bool
	Compress bool
	MaxItems uint
	Cleanup  bool
}

type Stats struct {
	NumberOfAddedItems uint
	NumberOfDupedItems uint
	NumberOfItems      uint
}

var DefaultOptions Options = Options{
	Dedupe:   true,
	Compress: false,
	MaxItems: math.MaxInt16,
	Cleanup:  true,
}
