package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var cfg *mysql.Config
var tmpl *template.Template

type PageData struct {
	Title          string
	Customers      []User
	CustomersStats []Customer
	Orders         []Order
	Pagination     Pagination
	Sort           string
	SearchQuery    string
}
type CustomersPageData struct {
	Title          string
	Customers      []User
	CustomersStats []Customer
	Orders         []Order
}

func initTemplates() {
	var err error
	tmpl = template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	})

	// Parse all HTML templates
	tmpl, err = tmpl.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
}

func main() {
	initTemplates()
	// Capture connection properties.
	cfg = mysql.NewConfig()
	cfg.User = os.Getenv("DBUSER")
	cfg.Passwd = os.Getenv("DBPASS")
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "ecommerce_db"

	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	http.HandleFunc("/", dashboardHandler)
	http.HandleFunc("/dashboard", dashboardHandler)

	// Add to master.go in the routes section
	http.HandleFunc("/orders", ordersHandler)
	http.HandleFunc("/order/new", createOrderHandler)
	http.HandleFunc("/order/create", createOrderHandler)

	http.HandleFunc("/order/cancel/", cancelOrderHandler)
	// Customer routes
	http.HandleFunc("/customers", customersHandler)
	http.HandleFunc("/customers/", customersHandler)
	http.HandleFunc("/customer/view/", viewCustomerHandler)
	http.HandleFunc("/customer/add", addCustomerHandler)
	http.HandleFunc("/customer/edit/", editCustomerHandler)
	http.HandleFunc("/customer/delete/", deleteCustomerHandler)
	http.HandleFunc("/customer/update/", updateCustomerHandler)

	// Product routes
	http.HandleFunc("/products", productsHandler)
	http.HandleFunc("/product/add", addProductHandler)
	http.HandleFunc("/product/create", createProductHandler)
	http.HandleFunc("/product/edit/", editProductHandler)
	http.HandleFunc("/product/update/", updateProductHandler)
	http.HandleFunc("/product/delete/", deleteProductHandler)

	http.HandleFunc("/order/view/", viewOrderHandler)
	http.HandleFunc("/order/update/", updateOrderHandler)

	http.HandleFunc("/replicate", replicateHandler)
	fmt.Println("Slave Node running on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
func replicateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Query    string `json:"query"`
		Database string `json:"database"`
		User     string `json:"user"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err) // Log error for debugging
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", reqBody.User, reqBody.Password, reqBody.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err) // Log connection error
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec(reqBody.Query)
	if err != nil {
		log.Printf("Failed to execute query: %v", err) // Log query execution error
		http.Error(w, "Failed to execute query", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Replication successful")
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Dashboard",
	}
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func replicateToMaster(query string, database string, user string, password string) {

	body := map[string]string{
		"query":    query,
		"database": database,
		"user":     user,
		"password": password,
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error encoding JSON: %v", err)
		return
	}

	resp, err := http.Post("http://192.168.39.28:8080/replicate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to replicate to master: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Master returned non-OK status: %v", resp.Status)
	}
}
