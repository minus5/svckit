package mdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"
)

func NewSnapshots(fs *Fs, dir string) *Snapshots {
	n := &Snapshots{
		fs:         fs,
		dir:        dir,
		pending:    make(map[int]int),
		fileLocks:  make(map[int]*sync.Mutex),
		closing:    make(chan struct{}),
		loopClosed: make(chan struct{}),
	}
	go n.loop()
	return n
}

// Snapshots priprema full na disku od pojedinacnih diffova.
// Periodicki i na Close ih spremi u mongo.
// Na Find potrazi na disku pa u mongo.
type Snapshots struct {
	fs         *Fs
	dir        string
	pending    map[int]int
	fileLocks  map[int]*sync.Mutex
	closing    chan struct{}
	loopClosed chan struct{}
	sync.Mutex
}

// Close clean exit.
// Zaustavi loop i posalje sve pending u mongo.
func (n *Snapshots) Close() {
	close(n.closing)
	<-n.loopClosed
	n.pushPending()
}

func (n *Snapshots) loop() {
	t := time.Tick(time.Minute * 8)
	for {
		select {
		case <-t:
			n.pushPending()
		case <-n.closing:
			close(n.loopClosed)
			return
		}
	}
}

// pushPending salje sve pending snapshots u mongo.
func (n *Snapshots) pushPending() {
	n.Lock()
	p := n.pending
	n.pending = make(map[int]int)
	n.Unlock()

	for id, version := range p {
		if version > 4 {
			n.push(id, version)
		}
	}
}

// push sprema snapshots s diska u mongo.
// Pronadje snapshot za taj id, version i ubaci ga u mongo GridFs.
func (n *Snapshots) push(id, version int) error {
	n.lock(id)
	defer n.unlock(id)
	// pronadji file
	fn := n.fn(id, version)
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	// posalji file u mongo
	typ := strconv.Itoa(id)
	uid := n.uid(id, version)
	if err := n.fs.Insert(typ, uid, time.Now(), f); err != nil {
		return err
	}
	// obrisi stare u mongo tako da ostane samo jedan
	n.fs.Compact(typ)
	//log.I("id", id).I("version", version).Debug("push")
	return nil
}

// fn naziv file-a na disku
func (n *Snapshots) fn(id, version int) string {
	return fmt.Sprintf("%s/%d-%d", n.dir, id, version)
}

// uid unique id za file u GridFs
func (n *Snapshots) uid(id, version int) string {
	return fmt.Sprintf("%d-%d", id, version)
}

// Find pronalazi cacheirane diffove.
// Prvo gleda na disk onda u mongo.
func (n *Snapshots) Find(id, version int, decode func(io.Reader) error) error {
	n.lock(id)
	defer n.unlock(id)

	if err := n.findDisk(id, version, decode); err == nil {
		//log.I("id", id).I("version", version).Debug("findDisk")
		return err
	}
	return n.findMongo(id, version, decode)
}

func (n *Snapshots) findMongo(id, version int, decode func(io.Reader) error) error {
	h := func(r io.ReadCloser) error {
		var buf bytes.Buffer
		tr := io.TeeReader(r, &buf)
		if err := readToDiff(tr, decode); err != nil {
			return err
		}
		if err := r.Close(); err != nil {
			return err
		}
		return n.create(id, version, &buf)
	}
	return n.fs.FindId(n.uid(id, version), h)
}

func (n *Snapshots) findDisk(id, version int, decode func(io.Reader) error) error {
	f, err := os.Open(n.fn(id, version))
	if err != nil {
		return err
	}
	defer f.Close()
	return readToDiff(f, decode)
}

// readToDiff raspakirava sadrzaj readera u array diff-ova.
// Prva cetri bajta su duzina diff-a, onda ide diff te duzine i tako citamo do EOF.
func readToDiff(f io.Reader, decode func(io.Reader) error) error {
	for {
		// procitaj duzinu 4 bajta
		sb := make([]byte, 4)
		n, err := f.Read(sb)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if n != 4 {
			return fmt.Errorf("unexpected read bytes %d", n)
		}
		// raspakuj duzinu u int
		size := int(binary.LittleEndian.Uint32(sb))
		// procitaj slijedecih size
		buf := make([]byte, size)
		n, err = f.Read(buf)
		if err != nil {
			return err
		}
		if n != size {
			return fmt.Errorf("unexpected read bytes %d", n)
		}
		// zovi dekodera s enkodiranim diff-om
		if err := decode(bytes.NewBuffer(buf)); err != nil {
			return err
		}
	}
	return nil
}

// create kreira novi file od sadrzaja r.
// Koristi se kada iz mongo spustamo na disk.
func (n *Snapshots) create(id, version int, r io.Reader) error {
	flag := os.O_CREATE | os.O_WRONLY
	f, err := os.OpenFile(n.fn(id, version), flag, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

// Append dodaje buf u disk cache.
// Pronadje na disku file od version-1, doda na njega buf i preimenuje ga u version.
func (n *Snapshots) Append(id, version int, buf []byte) error {
	n.lock(id)
	defer n.unlock(id)

	flag := os.O_CREATE | os.O_WRONLY
	if version > 1 {
		flag = flag | os.O_APPEND
	}
	fn := n.fn(id, version-1)
	f, err := os.OpenFile(fn, flag, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	sb := make([]byte, 4)
	binary.LittleEndian.PutUint32(sb, uint32(len(buf)))
	if _, err := f.Write(sb); err != nil {
		return err
	}
	if _, err := f.Write(buf); err != nil {
		return err
	}
	n.Lock()
	n.pending[id] = version // zapamti da na disku imamo snapshot koji bi mogli slati u mongo
	n.Unlock()
	//log.I("id", id).I("version", version).Debug("append")
	fn2 := n.fn(id, version)
	return os.Rename(fn, fn2)
}

// lock po fileu.
// Da nemamo jedan globalni radimo lock po svakom pojednom fileu.
func (n *Snapshots) lock(id int) {
	n.Lock()
	if l, ok := n.fileLocks[id]; ok {
		n.Unlock()
		l.Lock()
		return
	}
	var l sync.Mutex
	n.fileLocks[id] = &l
	n.Unlock()
	l.Lock()
}

func (n *Snapshots) unlock(id int) {
	n.Lock()
	l, ok := n.fileLocks[id]
	if !ok {
		l.Unlock()
		return
	}
	l.Unlock()
	delete(n.fileLocks, id)
	n.Unlock()
}
