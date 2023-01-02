//go:build 386

package disk

func init() {
	OpenPebbleDB = func(_ string) (DB, error) {
		return nil, ErrNotSupported
	}
}
