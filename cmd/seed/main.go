// seed creates an initial superadmin user.
// Usage: go run ./cmd/seed -email admin@example.com -password secret
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/johlun99/revio/internal/config"
	"github.com/johlun99/revio/internal/db"
)

func main() {
	email := flag.String("email", "", "Admin email (required)")
	password := flag.String("password", "", "Admin password (required)")
	role := flag.String("role", "superadmin", "Role: superadmin|admin|moderator")
	flag.Parse()

	if *email == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "usage: seed -email <email> -password <password> [-role <role>]")
		os.Exit(1)
	}

	cfg := config.Load()

	pool, err := db.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connect: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash password: %v\n", err)
		os.Exit(1)
	}

	var id string
	err = pool.QueryRow(context.Background(),
		`INSERT INTO admin_users (email, password_hash, role)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role
		 RETURNING id`,
		*email, string(hash), *role,
	).Scan(&id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "insert user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ admin user upserted: %s (%s) id=%s\n", *email, *role, id)
}
