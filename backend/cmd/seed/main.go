package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Get connection string from env or use default
	connStr := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASS")
	connStr += "@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME")

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Hash password for admin123
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	// Insert admin user
	_, err = db.Exec(
		"INSERT INTO admin_users (username, password_hash, email, rol, activo) VALUES (?, ?, ?, ?, ?)",
		"admin", string(hash), "admin@wsapi.local", "super_admin", true,
	)
	if err != nil {
		log.Fatalf("Error creating admin user: %v", err)
	}

	fmt.Println("Admin user created: admin / admin123")
}
