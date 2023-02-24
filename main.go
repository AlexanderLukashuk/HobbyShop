package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"fainal.net/api"
	"fainal.net/internal/jsonlog"
)

// postgres
//nurzhan - "postgres://postgres:admin@localhost/final_go?sslmode=disable"
//adiya   - os.Getenv("DSN") $env:DSN="postgres://postgres:20072004@localhost:5432/finalProject?sslmode=disable"
//Sasha   -

func main() {
	viper.SetConfigFile("ENV")
	viper.ReadInConfig()
	viper.AutomaticEnv()
	port := fmt.Sprint(viper.Get("PORT"))

	// port := os.Getenv("PORT")

	// if port == "" {
	// 	port = "3000"
	// }

	// portInt, _ := strconv.Atoi(port)

	var cfg api.Config
	flag.IntVar(&cfg.Port, "port", 8000, "API server port")
	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|staging|production)")

	// Read the DSN value from the db-dsn command-line flag into the config struct. We
	// default to using our development DSN if no flag is provided.
	// in powershell use next command: $env:DSN="postgres://postgres:20072004@localhost:5432/greenlight?sslmode=disable"
	// flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://postgres:admin@localhost/final_go?sslmode=disable", "PostgreSQL DSN")
	flag.StringVar(&cfg.Db.Dsn, "db-dsn", "postgres://postgres:postgres@localhost/final_go?sslmode=disable", "PostgreSQL DSN")

	// Setting restrictions on db connections
	flag.IntVar(&cfg.Db.MaxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.Db.MaxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.Db.MaxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max idle time")
	// flag.StringVar(&cfg.db.maxLifetime, "db-max-lifetime", "1h", "PostgreSQL max idle time")

	flag.Float64Var(&cfg.Limiter.Rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.Limiter.Burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.Limiter.Enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.Smtp.Host, "smtp-host", "smtp.office365.com", "SMTP host")
	flag.IntVar(&cfg.Smtp.Port, "smtp-port", 25, "SMTP port")
	// flag.StringVar(&cfg.smtp.username, "smtp-username", "211322@astanait.edu.kz", "SMTP username")
	flag.StringVar(&cfg.Smtp.Username, "smtp-username", "211437@astanait.edu.kz", "SMTP username")
	flag.StringVar(&cfg.Smtp.Password, "smtp-password", "Aitu2021!", "SMTP password")
	flag.StringVar(&cfg.Smtp.Sender, "smtp-sender", "Hobby Shop <211437@astanait.edu.kz>", "SMTP sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.Cors.TrustedOrigins = strings.Fields(val)
		return nil
	})

	r := mux.NewRouter().StrictSlash(true)

	flag.Parse()
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)

	// db, err := openDB(cfg)
	// if err != nil {
	// 	logger.PrintFatal(err, nil)
	// }

	// defer db.Close()
	// logger.PrintInfo("database connection pool established", nil)

	// app := &Application{
	// 	config: cfg,
	// 	logger: logger,
	// 	models: data.NewModels(db),
	// 	mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender), // data.NewModels() function to initialize a Models struct
	// }

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	// err = app.serve()
	// if err != nil {
	// 	logger.PrintFatal(err, nil)
	// }

	log.Println(http.ListenAndServe(":"+port, loggedRouter))

}

func openDB(cfg api.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.Db.Dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.Db.MaxIdleConns)
	db.SetMaxOpenConns(cfg.Db.MaxOpenConns)

	duration, err := time.ParseDuration(cfg.Db.MaxIdleTime)
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
