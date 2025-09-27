package main

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", "localhost", "prac", "prac", "prac")

	db, err := sql.Open("pgx", ps)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// ...
}
