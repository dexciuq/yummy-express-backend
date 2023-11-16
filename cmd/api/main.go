package main

import (
	"flag"
	"fmt"
	"github.com/dexciuq/yummy-express-backend/internal/jsonlog"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"sync"
)

const version = "1.0"

type config struct {
	port    int
	env     string
	limiter struct {
		enabled bool
		rps     float64
		burst   int
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	wg     sync.WaitGroup
}

func getEnvVar(key string) string {
	godotenv.Load()
	return os.Getenv(key)
}

func main() {
	var cfg config
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("Empty")
		port = "7000"
	}
	port_int, err := strconv.Atoi(port)
	flag.IntVar(&cfg.port, "port", port_int, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// Set up limitations for application
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}
