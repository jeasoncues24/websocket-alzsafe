package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"wsapi/internal/config"
	apihttp "wsapi/internal/http"
	"wsapi/internal/storage"
)

func main() {
	migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)

	flag.Usage = func() {
		fmt.Println("Usage: wsapi [OPTIONS] COMMAND")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  migrate status  Muestra el estado de las migraciones")
		fmt.Println("  migrate up      Aplica las migraciones pendientes")
		fmt.Println("  migrate down    Revierte la última migración")
		fmt.Println("  migrate adopt   Adopta el historial de golang-migrate en goforge")
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
		runMigrateCommand(migrateCmd)
		return
	}

	startServer()
}

func runMigrateCommand(migrateCmd *flag.FlagSet) {
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

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=America%%2FLima&multiStatements=true",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	runner := storage.NewMigrationRunner(dsn)
	ctx := context.Background()

	switch args[0] {
	case "status":
		runStatus(ctx, runner)
	case "up":
		runUp(ctx, runner)
	case "down":
		runDown(ctx, runner)
	case "adopt":
		runAdopt(ctx, runner)
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		migrateCmd.Usage()
	}
}

func runStatus(ctx context.Context, runner *storage.MigrationRunner) {
	entries, err := runner.Status(ctx)
	if err != nil {
		fmt.Printf("Error getting status: %v\n", err)
		os.Exit(1)
	}

	applied := 0
	for _, e := range entries {
		if e.Applied {
			applied++
		}
	}

	fmt.Printf("Migraciones: %d totales, %d aplicadas, %d pendientes\n\n",
		len(entries), applied, len(entries)-applied)

	for _, e := range entries {
		state := "pendiente"
		if e.Applied {
			state = "aplicada"
			if e.Dirty {
				state = "DIRTY"
			}
		}
		fmt.Printf("  [%06d] %-45s %s\n", e.Version, e.Name, state)
	}
}

func runUp(ctx context.Context, runner *storage.MigrationRunner) {
	fmt.Println("Aplicando migraciones pendientes...")

	if err := runner.RunMigrations(ctx); err != nil {
		fmt.Printf("Error running migrations: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Migraciones aplicadas.")
	fmt.Println("")
	runStatus(ctx, runner)
}

func runDown(ctx context.Context, runner *storage.MigrationRunner) {
	fmt.Println("Revirtiendo la última migración...")

	if err := runner.Rollback(ctx, 1); err != nil {
		fmt.Printf("Error rolling back: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Migración revertida.")
	fmt.Println("")
	runStatus(ctx, runner)
}

func runAdopt(ctx context.Context, runner *storage.MigrationRunner) {
	fmt.Println("Adoptando el historial de golang-migrate (schema_migrations) en goforge_migrations...")

	adopted, err := runner.Adopt(ctx)
	if err != nil {
		fmt.Printf("Error adopting migration history: %v\n", err)
		os.Exit(1)
	}

	if len(adopted) == 0 {
		fmt.Println("No había nada que adoptar (goforge_migrations ya estaba al día).")
		return
	}

	fmt.Printf("Marcadas como aplicadas %d migraciones: %v\n", len(adopted), adopted)
	fmt.Println("La tabla legacy schema_migrations se dejó intacta; elimínala manualmente cuando verifiques que todo funciona.")
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
