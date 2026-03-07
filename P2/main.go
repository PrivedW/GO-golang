package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Product struct {
	ID      int
	Model   string
	Company string
	Price   int
}

type DBInterface interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

func (s *Store) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.Query(query, args...)
}

type App struct {
	db DBInterface
}

func NewApp(db DBInterface) *App {
	return &App{db: db}
}

func (a *App) InsertProduct(model, company string, price int) (sql.Result, error) {
	return a.db.Exec("INSERT INTO products (model, company, price) VALUES (?, ?, ?)",
		model, company, price)
}

func (a *App) GetAllProducts() ([]Product, error) {
	rows, err := a.db.Query("SELECT model, price FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var model string
		var price int
		err := rows.Scan(&model, &price)
		if err != nil {
			return nil, err
		}
		products = append(products, Product{Model: model, Price: price})
	}
	return products, nil
}

func (a *App) DeleteByCompany(company string) (int64, error) {
	result, err := a.db.Exec("DELETE FROM products WHERE company = ?", company)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (a *App) Run() error {
	result, err := a.InsertProduct("Galaxy A52", "Samsung", 4300)
	if err != nil {
		return err
	}
	fmt.Printf("Inserted: %v\n", result)

	products, err := a.GetAllProducts()
	if err != nil {
		return err
	}

	fmt.Println("\nВсе модели и цены: ")
	for _, p := range products {
		fmt.Printf("%s: %d руб.\n", p.Model, p.Price)
	}

	count, err := a.DeleteByCompany("Samsung")
	if err != nil {
		return err
	}
	fmt.Printf("\nУдалено Samsung: %d\n", count)

	return nil
}

func main() {
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := NewStore(db)
	app := NewApp(store)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
