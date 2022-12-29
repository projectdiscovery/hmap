//go:build (arm || arm64) && windows

package disk

import (
	"bytes"
	"strconv"
	"sync"
	"time"
)

func init() {
	OpenPogrebDB = func(_ string) (DB, error) {
		return nil, ErrNotSupported
	}
}
