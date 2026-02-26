package cli

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/db/migrations"
	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/config"
	httpServer "github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/server/http"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

type CommandRunner struct {
	args        []string
	intFlags    intFlags
	stringFlags stringFlags
	boolFlags   boolFlags
}

func newCommandRunner(
	args []string,
	intFlags intFlags,
	stringFlags stringFlags,
	boolFlags boolFlags,
) *CommandRunner {
	return &CommandRunner{
		args:        args,
		intFlags:    intFlags,
		stringFlags: stringFlags,
		boolFlags:   boolFlags,
	}
}

func (cr *CommandRunner) runCommand(cmd string) error {
	switch cmd {
	case "help":
		cr.runHelp()
	case "demo":
		cr.runDemo()
	case "serve":
		cr.runServe()
	default:
		return errors.New("Unknown command")
	}
	return nil
}

func (cr *CommandRunner) runHelp() (exitCode int) {
	log.Print(helpText)
	return ExitOK
}

func (cr *CommandRunner) runDemo() (exitCode int) {
	cmd := exec.Command("docker", "--version")
	if err := cmd.Run(); err != nil {
		log.Print("docker is missing")
		return ExitUsage
	}

	dbPort, err := findFreePort()
	if err != nil {
		log.Printf("Failed to find a free port: %v", err)
		return ExitRuntime
	}

	dbContainerName := "ffg-demo-db"

	cmd = exec.Command("docker", "rm", "-f", dbContainerName)
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: failed to remove existing container: %v", err)
	}
	defer func() {
		if err := exec.Command("docker", "rm", "-f", dbContainerName).Run(); err != nil {
			log.Printf("Failed to stop %s container: %v", dbContainerName, err)
		}
	}()

	cmd = exec.Command(
		"docker", "run", "-d",
		"-p", fmt.Sprintf("%d:5432", dbPort),
		"-e", "POSTGRES_USER=demo",
		"-e", "POSTGRES_PASSWORD=demo",
		"-e", "POSTGRES_DB=demo",
		"--name", dbContainerName,
		"postgres:16-alpine",
	)

	log.Print("Starting a postgres container...")
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to start a container: %v", err)
		return ExitRuntime
	}
	log.Print("Container started")

	dbURL := fmt.Sprintf("postgres://demo:demo@localhost:%d/demo?sslmode=disable", dbPort)

	log.Print("Initializing database connection...")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Failed to initialize database connection: %v", err)
		return ExitRuntime
	}
	defer db.Close()

	retries := 10

	log.Print("Waiting for database to be ready...")
	for i := 1; i <= retries; i++ {
		if err = db.Ping(); err == nil {
			break
		}

		log.Printf("Database not ready (attempt %d/%d): %v", i, retries, err)
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Printf("Database not ready after %d retries: %v", retries, err)
		return ExitRuntime
	}
	log.Print("Database is ready")

	log.Print("Running migrations...")
	if err := runMigrations(dbURL); err != nil {
		log.Printf("Failed to apply migrations: %v", err)
		return ExitRuntime
	}

	log.Print("Loading config...")
	cfg, err := loadConfig(cr)
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return ExitRuntime
	}

	HTTPConfig := httpServer.HTTPConfig{
		Port:  cfg.Port,
	}

	server := httpServer.NewHTTPServer(HTTPConfig, db)

	log.Printf("Server starting on port %d...", cfg.Port)
	err = server.ListenAndServe()
	if err != nil {
		log.Printf("Failed to start http server on port %d: %v", HTTPConfig.Port, err)
		return ExitRuntime
	}

	return ExitOK
}

func (cr *CommandRunner) runServe() (exitCode int) {
	log.Print("Loading config...")
	cfg, err := loadConfig(cr)
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return ExitRuntime
	}

	log.Print("Initializing database connection...")
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Printf("Failed to initialize database connection: %v", err)
		return ExitRuntime
	}
	defer db.Close()
	log.Print("Waiting for database to be ready...")
	if err := db.Ping(); err != nil {
		log.Printf("Database is unavailable: %v", err)
		return ExitRuntime
	}
	log.Print("Database is ready")

	log.Print("Running migrations...")
	if err := runMigrations(cfg.DBURL); err != nil {
		log.Printf("Failed to apply migrations: %v", err)
		return ExitRuntime
	}

	HTTPConfig := httpServer.HTTPConfig{
		Port:  cfg.Port,
	}

	server := httpServer.NewHTTPServer(HTTPConfig, db)

	log.Printf("Server starting on port %d...", cfg.Port)
	err = server.ListenAndServe()
	if err != nil {
		log.Printf("Failed to start http server on port %d: %v", HTTPConfig.Port, err)
		return ExitRuntime
	}

	return ExitOK
}

func loadConfig(cr *CommandRunner) (config.Config, error) {
	path := *cr.stringFlags["config"]
	cfg, err := config.Load(path)
	if err != nil {
		return config.Config{}, errors.New(err.Error())
	}
	return cfg, nil

}

func runMigrations(dbURL string) error {
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to get migrations from file system: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		return fmt.Errorf("failed to set up migration: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}

func findFreePort() (port int, err error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	addr := ln.Addr().(*net.TCPAddr)
	port = addr.Port
	ln.Close()
	return port, nil
}

const helpText = `Feature Flag Gatekeeper (ffg)

Usage:
  ffg <command> [flags]

Commands:
  serve        Start the HTTP server
  demo         Run demo with local Postgres (Docker)
  help         Show this help message

Flags:
  --config     Path to config file
  --port       Port for the app (default: 8080)

Examples:
  ffg serve --config config.yaml
  ffg demo
`
