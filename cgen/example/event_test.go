package example

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurentMapIterationAndModify(t *testing.T) {
	e := Event{
		Home: "Hajduk",
		Away: "Dinamo",
		Markets: map[int]Market{
			0: Market{
				Name: "1x2",
				Outcomes: map[int]Outcome{
					0: Outcome{
						Name: "1",
						Odds: 1.25,
					},
					1: Outcome{
						Name: "2",
						Odds: 3.75,
					},
				},
			},
		},
	}

	odds := 1.25
	var wg sync.WaitGroup
	for i := 0; i < 8000; i++ {
		odds += 0.01
		d := EventDiff{}
		d.Markets.Empty(0).Outcomes.Empty(0).Odds = &odds

		e2 := e.Merge(d) // modifications on the data copy
		wg.Add(1)
		go func(e3 Event) {
			defer wg.Done()
			// map iterations
			for _, m := range e3.Markets {
				for _, o := range m.Outcomes {
					fmt.Printf("%f ", o.Odds)
				}
				fmt.Printf("\n")
			}
		}(e2)
	}
	wg.Wait()

	// original value is not modified
	assert.Equal(t, 1.25, e.Markets[0].Outcomes[0].Odds)
	assert.Equal(t, 3.75, e.Markets[0].Outcomes[1].Odds)
}

func TestDeepCopy(t *testing.T) {
	e := Event{
		Home: "Hajduk",
		Away: "Dinamo",
		Markets: map[int]Market{
			0: Market{
				Name: "1x2",
				Outcomes: map[int]Outcome{
					0: Outcome{
						Name: "1",
						Odds: 1.25,
					},
					1: Outcome{
						Name: "2",
						Odds: 3.75,
					},
				},
			},
		},
	}

	e2 := e.Copy()
	m := e2.Markets[0]
	m.Name = "12"
	e2.Markets[0] = m
	o := e2.Markets[0].Outcomes[0]
	o.Odds = 1.26
	e2.Markets[0].Outcomes[0] = o

	// e is unchanged
	assert.Equal(t, 1.25, e.Markets[0].Outcomes[0].Odds)
	assert.Equal(t, 3.75, e.Markets[0].Outcomes[1].Odds)
	assert.Equal(t, "1x2", e.Markets[0].Name)
	// e2 has new values
	assert.Equal(t, 1.26, e2.Markets[0].Outcomes[0].Odds)
	assert.Equal(t, 3.75, e2.Markets[0].Outcomes[1].Odds)
	assert.Equal(t, "12", e2.Markets[0].Name)
}
