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
  dbName := fmt.Sprintf("metaldocs_exec_%d", os.Getpid())
  admin, err := sql.Open("pgx", "host=127.0.0.1 port=5433 user=metaldocs_app password=Lepa12<>! dbname=postgres sslmode=disable")
  if err != nil { panic(err) }
  defer admin.Close()
  ctx := context.Background()
  if _, err := admin.ExecContext(ctx, "DROP DATABASE IF EXISTS \""+dbName+"\""); err != nil { panic(err) }
  if _, err := admin.ExecContext(ctx, "CREATE DATABASE \""+dbName+"\""); err != nil { panic(err) }
  db, err := sql.Open("pgx", "host=127.0.0.1 port=5433 user=metaldocs_app password=Lepa12<>! dbname="+dbName+" sslmode=disable")
  if err != nil { panic(err) }
  defer db.Close()
  baseline, err := os.ReadFile("db/baseline/0001_current_schema.sql")
  if err != nil { panic(err) }
  if _, err := db.ExecContext(ctx, string(baseline)); err != nil { fmt.Println("ERR:", err); os.Exit(1) }
  fmt.Println("OK")
}
