package disk

import (
	"io/ioutil"
	"os"
	"testing"
)

func utiltestGetPath(t *testing.T) (string, error) {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "hmaptest")
	if err != nil {
		t.Error(err)
	}
	return tmpdir, err
}

func utiltestRemoveDb(db DB, pbpath string) {
	db.Close()
	os.RemoveAll(pbpath)
}
