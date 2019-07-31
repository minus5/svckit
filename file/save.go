package file

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

func JSON(filename string, o interface{}) error {
	buf, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return Save(filename, buf)
}

func JSONPretty(filename string, o interface{}) error {
	buf, err := json.MarshalIndent(o, "  ", "  ")
	if err != nil {
		return err
	}
	return Save(filename, buf)
}

func Save(filename string, buf []byte) error {
	dir, _ := path.Split(filename)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, buf, 0644)
}
