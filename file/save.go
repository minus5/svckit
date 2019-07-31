package file

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

func saveJSON(filename string, o interface{}) error {
	buf, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return saveRaw(filename, buf)
}

func saveRaw(filename string, buf []byte) error {
	dir, _ := path.Split(filename)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, buf, 0644)
}
