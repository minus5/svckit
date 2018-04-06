// Package saga defines simple saga orcestrator
// - What is saga: https://www.youtube.com/watch?v=0UTOLRTwOX0
package saga

// Saga contains all steps saga must perform
type Saga struct {
	steps        []Step  // forward steps, executed first
	notify       FStep   // notifications step, executed after forward steps
	compensating []Step  // compensating steps, added when forward step fails, executed after steps and notify
	cleanup      []FStep // cleanup steps, always executed last
}

// Step defines step interface for forward / cleanup steps
//
// NOTE: error returned from Do or Compensate stops saga execution
type Step interface {
	Do() error         // Do executes forward step
	Finished() bool    // Finished returns true when Do is finished
	Successful() bool  // Successful returns true when Do was successful
	Failed() bool      // Failed returns true Do failed
	Compensate() error // Compensate executes compensate step (when Do failed)
	Compensated() bool // Compensate returns true when succesful
}

// FStep is forward only step, no compensate option
//
// NOTE: used as notify or cleanup step
type FStep interface {
	Do(bool) error  // Do executes forward step, bool param is saga success flag
	Finished() bool // Finished returns true when Do is finished
}

// New creates new saga
// - steps   - forward steps to execute
// - notify  - notify step executed when result of forward steps
//             is known regardless of success or fail, executed
//             before compensate and cleanup steps
// - cleanup - steps always executed at the end of saga
//             regardless of success or fail
//
// NOTE: error in any step stops execution
func New(steps []Step, notify FStep, cleanup []FStep) *Saga {
	return &Saga{
		steps:        steps,
		notify:       notify,
		cleanup:      cleanup,
		compensating: make([]Step, 0),
	}
}

// Do executes saga, this should be called to start saga execution
// - any error returned stops saga execution
// - it is callers responsibility to decide what to do in case of error returned
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

// doForward executes forward steps by executing Do for each step
// adds compensating steps in case execution fails
//
// Returns:
// - success - status of forward step execution
// - error   - when step fails with error
//
// NOTE: compensating steps are added only when step Do is not failed
func (s *Saga) doForward() (bool, error) {
	success := true
	for _, step := range s.steps {
		if !step.Finished() {
			if err := step.Do(); err != nil {
				return false, err
			}
		}
		if !step.Failed() { // ako je aborted ili successful
			s.compensating = append(s.compensating, step)
		}
		if !step.Successful() { // ako je aborted ili failed
			success = false
			break
		}
	}
	return success, nil
}

// doCompensating executes Compensate for each succesful step
// in reverse order
//
// Returns error when Compensate fails
//
// NOTE: this is executed only when forward step
// Succesful returns false
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

// doCleanup executes all cleanup steps
//
// Returns error when cleanup Do fails
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
