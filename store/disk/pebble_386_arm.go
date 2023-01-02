//go:build 386 || arm

package disk

func init() {
	OpenPebbleDB = func(_ string) (DB, error) {
		return nil, ErrNotSupported
	}
}
