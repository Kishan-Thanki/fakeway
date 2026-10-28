package server

import (
	"flag"
	"time"
)

const (
	DEFAULT_ADDR          = ":8080"
	DEFAULT_READ_TIMEOUT  = 10 * time.Second
	DEFAULT_WRITE_TIMEOUT = 35 * time.Second
	DEFAULT_IDLE_TIMEOUT  = 120 * time.Second
)

type config struct {
	addr         string
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
}

func parseFlags() config {
	var cfg config

	flag.StringVar(
		&cfg.addr,
		"addr",
		DEFAULT_ADDR,
		"HTTP server address",
	)

	flag.DurationVar(
		&cfg.readTimeout,
		"read-timeout",
		DEFAULT_READ_TIMEOUT,
		"Maximum request read duration",
	)

	flag.DurationVar(
		&cfg.writeTimeout,
		"write-timeout",
		DEFAULT_WRITE_TIMEOUT,
		"Maximum response write duration",
	)

	flag.DurationVar(
		&cfg.idleTimeout,
		"idle-timeout",
		DEFAULT_IDLE_TIMEOUT,
		"Maximum idle connection duration",
	)

	flag.Parse()

	return cfg
}
