/* package saga definira jednostavan saga orkestrator
Sto je saga: https://www.youtube.com/watch?v=0UTOLRTwOX0
*/
package saga

type Saga struct {
	steps        []Step
	compensating []Step
	notify       FStep
	cleanup      []FStep
}

// Step definira metofde koje ima korak u sagi
// error u Do ili Compensate zaustavlja sagu i vraca taj error iz saga.Do
type Step interface {
	Do() error         // izvrsi forward akcije
	Compensate() error // izvrsi compensate/cancel/storno
	Finished() bool    // je li Do metoda odradjena
	Successful() bool  // Do je bio uspjesan
	Failed() bool      // Do je zavrsio ali neuspjesno
	Aborted() bool     // Do nije zavrsio, recimo timeout
	Compensated() bool // Compensate je uspjesno zavrsio
}

// FStep Forward only Step
// Korak koji moze ici samo prema naprijed, nema Compensate dijela
type FStep interface {
	Do(bool) error  // Izvrsi forward akciju, bool parametar uspjesnost sage
	Finished() bool // je li Do metoda odradjena
}

// New krira novu sagu
//     steps   - koraci koje treba izvrsiti
//     notify  - poziva se kada znamo rezultat sage, je li uspjesan ili neuspjesan zavrsetak
//               zove se prije compensate koraka i prije cleanup
//               cim znamo koji je rezultat
//     cleanup - koraci koje izvrsavamo uvijek na kraju
//               bilo da je rezultat sage uspjesan ili ne
func New(steps []Step, notify FStep, cleanup []FStep) *Saga {
	return &Saga{
		steps:        steps,
		notify:       notify,
		cleanup:      cleanup,
		compensating: make([]Step, 0),
	}
}

// Do vraca nil ako je saga dogurala do kraja
// error ako smo se negdje po putu raspali
// client-ova je odgovornost odluciti po tipu greske da li sagu ponovo pokrenuti
func (s *Saga) Do() error {
	success, err := s.doForward()
	if err != nil {
		return err
	}
	if !s.notify.Finished() {
		if err := s.notify.Do(success); err != nil {
			return err
		}
	}
	if !success {
		if err := s.doCompensating(); err != nil {
			return err
		}
	}
	if err := s.doCleanup(success); err != nil {
		return err
	}
	return nil
}

// doForward izvrsava Do za svaki korak
// Vraca:
//       success - ako su svi koraci bili uspjesni
//       error   - ako je neki od koraka zavriso greskom
// Priprema listu compensating koraka,
// one koje treba kompnezirati u slucaju neuspjesnog zavrsetka.
func (s *Saga) doForward() (bool, error) {
	success := true
	for _, step := range s.steps {
		if !step.Finished() {
			if err := step.Do(); err != nil {
				return false, err
			}
		}
		if !step.Failed() { // ako je aborted ili successful
			s.addCompensating(step)
		}
		if !step.Successful() { // ako je aborted ili failed
			success = false
			break
		}
	}
	return success, nil
}

func (s *Saga) addCompensating(step Step) {
	s.compensating = append(s.compensating, step)
}

// doCompensating poziva Compensate za svaki zapoceti korak
// Vraca error ako neki od koraka nije uspjesan.
func (s *Saga) doCompensating() error {
	for i := len(s.compensating) - 1; i >= 0; i-- {
		step := s.compensating[i]
		if step.Compensated() {
			continue
		}
		err := step.Compensate()
		if err != nil {
			return err
		}
	}
	return nil
}

// doCleanup poizva Do za cleanup korake
// Vraca error ako neki od koraka nije uspjesan.
func (s *Saga) doCleanup(success bool) error {
	for _, step := range s.cleanup {
		if step.Finished() {
			continue
		}
		if err := step.Do(success); err != nil {
			return err
		}
	}
	return nil
}
