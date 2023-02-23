package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"fainal.net/internal/data"
	"fainal.net/internal/jsonlog"
	"fainal.net/internal/mailer"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

// postgres
//nurzhan - "postgres://postgres:admin@localhost/final_go?sslmode=disable"
//adiya   - os.Getenv("DSN") $env:DSN="postgres://postgres:20072004@localhost:5432/finalProject?sslmode=disable"
//Sasha   -

func main() {
	// fiber.SetConfigFile("ENV")
	// viper.ReadInConfig()
	// viper.AutomaticEnv()
	// port := fmt.Sprint(viper.Get("PORT"))

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	portInt, _ := strconv.Atoi(port)

	var cfg config
	flag.IntVar(&cfg.port, "port", portInt, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// Read the DSN value from the db-dsn command-line flag into the config struct. We
	// default to using our development DSN if no flag is provided.
	// in powershell use next command: $env:DSN="postgres://postgres:20072004@localhost:5432/greenlight?sslmode=disable"
	// flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://postgres:admin@localhost/final_go?sslmode=disable", "PostgreSQL DSN")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://postgres:postgres@localhost/final_go?sslmode=disable", "PostgreSQL DSN")

	// Setting restrictions on db connections
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max idle time")
	// flag.StringVar(&cfg.db.maxLifetime, "db-max-lifetime", "1h", "PostgreSQL max idle time")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.office365.com", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	// flag.StringVar(&cfg.smtp.username, "smtp-username", "211322@astanait.edu.kz", "SMTP username")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "211437@astanait.edu.kz", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "Aitu2021!", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Hobby Shop <211437@astanait.edu.kz>", "SMTP sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender), // data.NewModels() function to initialize a Models struct
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)

	if err != nil {
		return nil, err
	}

	return db, nil
}
