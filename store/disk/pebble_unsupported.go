//go:build 386 || arm || openbsd

package disk

func init() {
	OpenPebbleDB = func(_ string) (DB, error) {
		return nil, ErrNotSupported
	}
}
