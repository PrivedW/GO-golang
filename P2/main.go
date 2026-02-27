package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type product struct {
	id      int
	model   string
	company string
	price   int
}

func main() {

	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	result, err := db.Exec("INSERT INTO products (model, company, price) VALUES (?, ?, ?)",
		"Galaxy A52", "Samsung", 4300)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nВсе модели и цены: ")
	rows, err := db.Query("SELECT model, price FROM products")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var model string
		var price int
		err := rows.Scan(&model, &price)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("%s: %d руб.\n", model, price)
	}

	result, err = db.Exec("DELETE FROM products WHERE company = ?", "Samsung")
	if err != nil {
		panic(err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("\nУдалено Samsung: %d\n", rowsAffected)
}
