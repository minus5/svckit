package util

import (
	"log"
	"testing"
	"time"
)

// ovo je vise primjer koristenja nego test
// stavljam skip jer traje neko vrijeme
// TODO - napravi nesto inteligentnije
func TestOneAtTheTime(t *testing.T) {
	t.Skip()

	var mu OneAtTheTime

	once := func() bool {
		return mu.Do(func() {
			time.Sleep(1e9)
		})
	}

	for i := 1; i < 50; i++ {
		go func(i int) {
			log.Printf("%d ok: %v", i, once())
		}(i)
		time.Sleep(1e8)
	}

}
