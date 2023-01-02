//go:build 386

package disk

import (
	"bytes"
	"strconv"
	"sync"
	"time"
)

func init() {
	OpenPebbleDB = func(_ string) (DB, error) {
		return nil, ErrNotSupported
	}
}
