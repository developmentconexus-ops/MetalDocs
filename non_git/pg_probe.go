//go:build ignore

package main

import (
  "context"
  "database/sql"
  "fmt"
  "os"
  _ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
  host := os.Getenv("PGHOST")
  port := os.Getenv("PGPORT")
  dbname := os.Getenv("PGDATABASE")
  user := os.Getenv("PGUSER")
  pass := os.Getenv("PGPASSWORD")
  sslmode := os.Getenv("PGSSLMODE")
  dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", host, port, dbname, user, pass, sslmode)
  db, err := sql.Open("pgx", dsn)
  if err != nil {
    fmt.Println("open error:", err)
    os.Exit(1)
  }
  defer db.Close()
  if err := db.PingContext(context.Background()); err != nil {
    fmt.Println("ping error:", err)
    os.Exit(1)
  }
  fmt.Println("postgres credential valid")
}
