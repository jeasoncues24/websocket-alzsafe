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
	flag.Usage = func() {
		fmt.Println("Uso: wsapi [COMANDO]")
		fmt.Println("")
		fmt.Println("Sin comando: inicia el servidor HTTP.")
		fmt.Println("")
		fmt.Println("Comandos:")
		fmt.Println("  serve                     Inicia el servidor HTTP (igual que sin comando)")
		fmt.Println("  migrate status | status   Muestra el estado de las migraciones")
		fmt.Println("  migrate up     | up       Aplica las migraciones pendientes")
		fmt.Println("  migrate down   | down     Revierte la última migración")
		fmt.Println("  migrate adopt  | adopt    Adopta el historial de golang-migrate en goforge")
		fmt.Println("")
		fmt.Println("Opciones:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 1 {
		startServer()
		return
	}

	switch flag.Arg(0) {
	case "serve", "server":
		startServer()
	case "migrate", "migration":
		runMigrateCommand(flag.Args()[1:])
	case "status", "up", "down", "adopt":
		// Atajos: equivalen a "migrate <comando>".
		runMigrateCommand(flag.Args())
	default:
		fmt.Printf("Comando desconocido: %q\n\n", flag.Arg(0))
		flag.Usage()
		os.Exit(1)
	}
}

func runMigrateCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Falta el subcomando de migración (status | up | down | adopt)")
		fmt.Println("")
		flag.Usage()
		os.Exit(1)
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
		fmt.Printf("Subcomando de migración desconocido: %q\n\n", args[0])
		flag.Usage()
		os.Exit(1)
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
