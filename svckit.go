package svckit

import (
	"expvar"
	"github.com/mnu5/svckit/env"
	"time"
)

var (
	startTime = time.Now()
)

func init() {
	publishExpvar()
}

func publishExpvar() {
	expvar.Publish("svckit.stats", expvar.Func(func() interface{} {
		stats := struct {
			Start    time.Time     `json:"start"`
			Uptime   time.Duration `json:"uptime"`
			AppName  string        `json:"appName"`
			Hostname string        `json:"hostname"`
		}{
			Start:    startTime,
			AppName:  env.AppName(),
			Hostname: env.Hostname(),
			Uptime:   time.Now().Sub(startTime) / time.Second,
		}
		return stats
	}))
}
