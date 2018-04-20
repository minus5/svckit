package mdb

import (
	"sync"

	"github.com/globalsign/mgo"
)

const (
	colSequences = "sequences"
	metricKey    = "sequence"
)

// Sequence represents a ... sequence.
type Sequence struct {
	sync.Mutex
	db        *Mdb
	key       string
	next      uint64
	leased    uint64
	bandwidth uint64
}

func (db *Mdb) GetSequence(key string, bandwidth uint64) (*Sequence, error) {
	seq := &Sequence{
		db:        db,
		key:       key,
		next:      0,
		leased:    0,
		bandwidth: bandwidth,
	}
	err := seq.updateLease()
	return seq, err
}

// Next would return the next integer in the sequence, updating the lease by running a transaction
// if needed.
func (seq *Sequence) Next() (uint64, error) {
	seq.Lock()
	defer seq.Unlock()

	if seq.next >= seq.leased {
		if err := seq.updateLease(); err != nil {
			return 0, err
		}
	}
	val := seq.next
	seq.next++
	return val, nil
}

// Release the leased sequence to avoid wasted integers. This should be done right
// before closing the associated DB. However it is valid to use the sequence after
// it was released, causing a new lease with full bandwidth.
func (seq *Sequence) Release() error {
	seq.Lock()
	defer seq.Unlock()

	err := seq.update(seq.next)
	seq.leased = seq.next
	return err
}

type mgoSequence struct {
	Id     string `bson:"_id"`
	Leased uint64
}

func (seq *Sequence) updateLease() error {
	if seq.leased == 0 {
		leased, err := seq.find()
		if err == mgo.ErrNotFound {
			leased := seq.leased + seq.bandwidth
			if err := seq.create(leased); err != nil {
				return err
			}
			seq.leased = leased
			seq.next = 1
			return nil
		}
		if err != nil {
			return err
		}

		seq.leased = leased
		seq.next = seq.leased
		return nil
	}

	leased := seq.leased + seq.bandwidth
	if err := seq.update(leased); err != nil {
		return err
	}
	seq.leased = leased
	return nil
}

func (seq *Sequence) update(leased uint64) error {
	return seq.db.UseSafe(colSequences, metricKey, func(col *mgo.Collection) error {
		current := &mgoSequence{Id: seq.key, Leased: seq.leased}
		pending := &mgoSequence{Id: seq.key, Leased: leased}
		return col.Update(current, pending)
	})
}

func (seq *Sequence) find() (uint64, error) {
	var leased uint64
	err := seq.db.UseSafe(colSequences, metricKey, func(col *mgo.Collection) error {
		ms := &mgoSequence{Id: seq.key}
		err := col.FindId(seq.key).One(ms)
		if err != nil {
			return err
		}
		leased = ms.Leased
		return nil
	})
	return leased, err
}

func (seq *Sequence) create(leased uint64) error {
	return seq.db.UseSafe(colSequences, metricKey, func(col *mgo.Collection) error {
		ms := &mgoSequence{Id: seq.key, Leased: leased}
		return col.Insert(ms)
	})
}
