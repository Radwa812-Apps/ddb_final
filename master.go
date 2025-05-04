package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB
var tmpl *template.Template
var cfg *mysql.Config

type PageData struct {
	Title          string
	Customers      []User
	CustomersStats []Customer
	Orders         []Order
	Pagination     Pagination
}
type CustomersPageData struct {
	Title          string
	Customers      []User
	CustomersStats []Customer
	Orders         []Order
}

func main() {

	tmpl = template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	})

	// Load templates
	tmpl = template.Must(tmpl.ParseGlob("templates/*.html"))

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

	// In master.go, remove the duplicate route registration
	http.HandleFunc("/", dashboardHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/database/", databaseTablesHandler)
	//http.HandleFunc("/tables", tablesHandler)
	http.HandleFunc("/database/auth", databaseAuthHandler)
	//http.HandleFunc("/table/columns/", tableColumnsHandler)
	//http.HandleFunc("/table/edit/", editTableHandler)
	//http.HandleFunc("/table/delete/", deleteTableHandler)
	//http.HandleFunc("/table/update/", updateTableHandler)

	// Database management routes
	http.HandleFunc("/databases", databasesHandler)
	http.HandleFunc("/database/create", createDatabaseHandler)
	http.HandleFunc("/database/add", addDatabaseHandler)        // Show add form
	http.HandleFunc("/database/delete/", deleteDatabaseHandler) // Handle delete
	http.HandleFunc("/tables", tablesHandler)                   // List tables
	http.HandleFunc("/table/add", addTableHandler)              // Show add form
	http.HandleFunc("/table/create", createTableHandler)        // Handle form submission
	http.HandleFunc("/table/edit/", editTableHandler)           // Show edit form
	http.HandleFunc("/table/update/", updateTableHandler)       // Handle edit form submission
	http.HandleFunc("/table/delete/", deleteTableHandler)       // Handle delete
	http.HandleFunc("/table/columns/", tableColumnsHandler)     // Show columns
	http.HandleFunc("/column/add/", addColumnHandler)           // Show add column form
	http.HandleFunc("/column/create/", createColumnHandler)     // Handle column creation
	http.HandleFunc("/column/delete/", deleteColumnHandler)     // Handle column deletion

	// Add to master.go in the routes section
	http.HandleFunc("/orders", ordersHandler)
	http.HandleFunc("/order/new", createOrderHandler)
	http.HandleFunc("/order/create", createOrderHandler)

	http.HandleFunc("/order/cancel/", cancelOrderHandler)
	http.HandleFunc("/customers/", customersHandler)

	// Customer routes
	http.HandleFunc("/customer/add", addCustomerHandler)
	http.HandleFunc("/customers/add", addCustomerSubmitHandler)
	http.HandleFunc("/customer/edit/", editCustomerHandler)
	http.HandleFunc("/customer/delete/", deleteCustomerHandler)
	http.HandleFunc("/customer/update/", updateCustomerHandler)

	// Product routes
	http.HandleFunc("/products", productsHandler)
	http.HandleFunc("/product/add", addProductHandler)        // Show add form
	http.HandleFunc("/product/create", createProductHandler)  // Handle form submission
	http.HandleFunc("/product/edit/", editProductHandler)     // Show edit form
	http.HandleFunc("/product/update/", updateProductHandler) // Handle edit form submission
	http.HandleFunc("/product/delete/", deleteProductHandler) // Handle delete

	// Remove this duplicate line:
	// http.HandleFunc("/order/create", createOrderHandler)
	http.HandleFunc("/order/view/", viewOrderHandler)
	http.HandleFunc("/order/update/", updateOrderHandler)

	// // Routes for displaying the form
	// http.HandleFunc("/table/create", addTableFormHandler)

	// // Route for processing form submission
	// http.HandleFunc("/tables/create", addTableSubmitHandler)

	//replicaate routes
	http.HandleFunc("/replicate", replicateHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Master running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

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
func replicateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Query string `json:"query"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(reqBody.Query)
	if err != nil {
		http.Error(w, "Failed to execute query", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Replication successful")
}
