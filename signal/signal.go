package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cenkalti/backoff"
)

func Term() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}

func Usr1() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	return c
}

func WaitForInterupt() {
	c := make(chan os.Signal, 1)
	//SIGINT je ctrl-C u shell-u, SIGTERM salje upstart kada se napravi sudo stop ...
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

// WithExponentialBackoff will retry handler on each error.
// Retries are in exponentialy increasing interval.
// With max interval between retries of 10 seconds, and max elapsed time of 1 minute.
func WithExponentialBackoff(op func() error) error {
	b := backoff.NewExponentialBackOff()
	b.MaxInterval = 10 * time.Second   //max interval between retries
	b.MaxElapsedTime = 1 * time.Minute //how log will we retry
	return backoff.Retry(op, b)
}

func WithBackoff(ctx context.Context, op func() error,
	maxInterval, maxElapsedTime time.Duration) error {
	b := backoff.NewExponentialBackOff()
	b.MaxInterval = maxInterval
	b.MaxElapsedTime = maxElapsedTime
	bc := backoff.WithContext(b, ctx)
	return backoff.Retry(op, bc)
}

// InteruptContext returns context which will be closed on application interupt
func InteruptContext() context.Context {
	ctx, stop := context.WithCancel(context.Background())
	go func() {
		WaitForInterupt()
		stop()
	}()
	return ctx
}
