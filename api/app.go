package api

import (
	"sync"

	"fainal.net/internal/data"
	"fainal.net/internal/jsonlog"
	"fainal.net/internal/mailer"
)

const Version = "1.0.0"

type Application struct {
	config Config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

type Config struct {
	Port int
	Env  string
	Db   struct {
		Dsn          string
		MaxOpenConns int
		MaxIdleConns int
		MaxIdleTime  string
	}
	Limiter struct {
		Rps     float64
		Burst   int
		Enabled bool
	}
	Smtp struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}
	Cors struct {
		TrustedOrigins []string
	}
}
