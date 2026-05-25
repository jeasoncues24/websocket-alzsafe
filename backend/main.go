package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"wsapi/internal/config"
	apihttp "wsapi/internal/http"
	"wsapi/internal/storage"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)

	verbose := flag.Bool("v", false, "verbose output")
	flag.Usage = func() {
		fmt.Println("Usage: wsapi [OPTIONS] COMMAND")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  migrate status  Show applied migrations")
		fmt.Println("  migrate up      Run pending migrations")
		fmt.Println("  migrate down    Revert last migration")
		fmt.Println("")
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		startServer()
		return
	}

	if flag.Arg(0) == "migrate" || flag.Arg(0) == "migration" {
		runMigrateCommand(migrateCmd, verbose)
		return
	}

	startServer()
}

func runMigrateCommand(migrateCmd *flag.FlagSet, verbose *bool) {
	args := flag.Args()[1:]

	if len(args) < 1 {
		migrateCmd.Usage()
		return
	}

	cfg := config.Load()
	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBName == "" || cfg.DBUser == "" {
		fmt.Println("Error: Database not configured. Set DB_HOST, DB_PORT, DB_NAME, DB_USER in .env")
		os.Exit(1)
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=America%%2FLima&multiStatements=true",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName))
	if err != nil {
		fmt.Printf("Error: Cannot connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("Error: Cannot ping database: %v\n", err)
		os.Exit(1)
	}

	runner := storage.NewMigrationRunner()

	switch args[0] {
	case "status":
		runStatus(runner, db, *verbose)
	case "up":
		runUp(runner, db, cfg.DBName, *verbose)
	case "down":
		runDown(runner, db, *verbose)
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		migrateCmd.Usage()
	}
}

func runStatus(runner *storage.MigrationRunner, db *sql.DB, verbose bool) {
	migrations, err := runner.GetAppliedMigrations(db)
	if err != nil {
		fmt.Printf("Error getting migrations: %v\n", err)
		os.Exit(1)
	}

	version, err := runner.GetCurrentVersion(db)
	if err != nil {
		fmt.Printf("Error getting version: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Current version: %d\n", version)
	fmt.Printf("Applied migrations: %d\n", len(migrations))
	fmt.Println("")

	if len(migrations) == 0 {
		fmt.Println("No migrations applied")
		return
	}

	fmt.Println("Migrations:")
	for _, m := range migrations {
		fmt.Printf("  [%d] %s - %s\n", m.Version, m.Description, m.AppliedAt)
	}
}

func runUp(runner *storage.MigrationRunner, db *sql.DB, dbName string, verbose bool) {
	fmt.Println("Running migrations...")

	// Force fresh connection settings
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Set session variables
	db.Exec("SET autocommit = 1")
	db.Exec("SET unique_checks = 1")
	db.Exec("SET foreign_key_checks = 1")

	err := runner.RunMigrations(db)
	if err != nil {
		fmt.Printf("Error running migrations: %v\n", err)
		os.Exit(1)
	}

	// Verify tables after
	var count int
	db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name != 'schema_migrations'", dbName).Scan(&count)
	fmt.Printf("Tables created (excluding schema_migrations): %d\n", count)

	version, _ := runner.GetCurrentVersion(db)
	fmt.Printf("Migrations applied. Current version: %d\n", version)
}

func runDown(runner *storage.MigrationRunner, db *sql.DB, verbose bool) {
	fmt.Println("Reverting last migration...")

	err := runner.Rollback(db)
	if err != nil {
		fmt.Printf("Error rolling back: %v\n", err)
		os.Exit(1)
	}

	version, _ := runner.GetCurrentVersion(db)
	fmt.Printf("Migration reverted. Current version: %d\n", version)
}

func startServer() {
	cfg := config.Load()
	port := cfg.AppPort
	if port == "" {
		fmt.Println("Error: APP_PORT not configured. Set APP_PORT in .env")
		os.Exit(1)
	}
	fmt.Printf("Servidor WhatsApp API iniciado en el puerto %s\n", port)
	router := apihttp.NewRouter()

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("error exponiendo puerto %s: %v", port, err)
		os.Exit(1)
	}

	type startupRunner interface {
		RunStartupTasks(context.Context)
	}
	if runner, ok := router.(startupRunner); ok {
		runner.RunStartupTasks(context.Background())
	}

	if err := http.Serve(listener, router); err != nil {
		fmt.Printf("error iniciando servidor: %v", err)
	}
}
