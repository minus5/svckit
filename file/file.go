package file

import (
	"compress/gzip"
	"io"
	"os"
	"path"
)

type writer struct {
	path       string
	file       *os.File
	gzipWriter *gzip.Writer
	gz         bool
}

func NewWriter(path string) (io.WriteCloser, error) {
	return newWriter(path, false)
}

func NewGzWriter(path string) (io.WriteCloser, error) {
	return newWriter(path, true)
}

func newWriter(path string, gz bool) (io.WriteCloser, error) {
	w := &writer{
		path: path,
		gz:   gz,
	}
	if err := w.mkdirAll(); err != nil {
		return nil, err
	}
	if err := w.open(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *writer) open() error {
	file, err := os.Create(w.path)
	if err != nil {
		return err
	}
	w.file = file
	if w.gz {
		w.gzipWriter = gzip.NewWriter(file)
	}
	return nil
}

func (w *writer) mkdirAll() error {
	dir, _ := path.Split(w.path)
	return os.MkdirAll(dir, os.ModePerm)
}

func (w *writer) Write(p []byte) (int, error) {
	if w.gz {
		n, err := w.gzipWriter.Write(p)
		if err == nil {
			w.gzipWriter.Flush()
			w.file.Sync()
		}
		return n, err
	}
	return w.file.Write(p)
}

func (w *writer) Close() error {
	if w.gz {
		if err := w.gzipWriter.Flush(); err != nil {
			return err
		}
		if err := w.gzipWriter.Close(); err != nil {
			return err
		}
	}
	return w.file.Close()
}

type reader struct {
	path       string
	file       *os.File
	gzipReader *gzip.Reader
	gz         bool
}

func NewReader(path string) (io.ReadCloser, error) {
	return newReader(path, false)
}

func NewGzReader(path string) (io.ReadCloser, error) {
	return newReader(path, true)
}

func newReader(path string, gz bool) (io.ReadCloser, error) {
	r := &reader{
		path: path,
		gz:   gz,
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	r.file = file
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	r.gzipReader = gzipReader
	return r, nil
}

func (r *reader) Read(p []byte) (int, error) {
	if r.gz {
		return r.gzipReader.Read(p)
	}
	return r.file.Read(p)
}

func (r *reader) Close() error {
	if r.gz {
		if err := r.gzipReader.Close(); err != nil {
			return err
		}
	}
	return r.file.Close()
}
