package filekv

import "math"

var (
	BufferSize = 50 * 1024 * 1024 // 50Mb
	Separator  = ";;;"
	NewLine    = "\n"
)

type Options struct {
	Path           string
	Dedupe         bool
	Compress       bool
	MaxItems       uint
	Cleanup        bool
	SkipEmpty      bool
	FilterCallback func(k, v []byte) bool
}

type Stats struct {
	NumberOfFilteredItems uint
	NumberOfAddedItems    uint
	NumberOfDupedItems    uint
	NumberOfItems         uint
}

var DefaultOptions Options = Options{
	Dedupe:   true,
	Compress: false,
	MaxItems: math.MaxInt16,
	Cleanup:  true,
}
