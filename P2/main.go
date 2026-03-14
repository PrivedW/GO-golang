package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type Product struct {
	ID      int    `json:"id"`
	Model   string `json:"model"`
	Company string `json:"company"`
	Price   int    `json:"price"`
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
	rows, err := a.db.Query("SELECT id, model, company, price FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Model, &p.Company, &p.Price)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
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

func (a *App) createProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var product Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := a.InsertProduct(product.Model, product.Company, product.Price)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	product.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func (a *App) getAllProductsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	products, err := a.GetAllProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (a *App) deleteProductsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	company := r.URL.Query().Get("company")
	if company == "" {
		http.Error(w, "Company parameter required", http.StatusBadRequest)
		return
	}

	count, err := a.DeleteByCompany(company)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"deleted_count": count,
		"company":       company,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *App) Run() error {
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			a.getAllProductsHandler(w, r)
		case http.MethodPost:
			a.createProductHandler(w, r)
		case http.MethodDelete:
			a.deleteProductsHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return http.ListenAndServe(":8080", nil)
}

func main() {
	fmt.Fprintln(os.Stdout, "СЕРВЕР ЗАПУЩЕН НА ПОРТУ 8080")
	os.Stdout.Sync()

	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка открытия БД: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		model TEXT,
		company TEXT,
		price INTEGER
	)`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания таблицы: %v\n", err)
		os.Exit(1)
	}

	store := NewStore(db)
	app := NewApp(store)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка сервера: %v\n", err)
		os.Exit(1)
	}
}
