package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	_ "github.com/lib/pq"
)

// Snaps represents the list of slave nodes
type Snap struct {
	Name     string
	Address  string
	Port     string
	Username string
}
type Pagination struct {
	CurrentPage int
	PerPage     int
	TotalItems  int
	TotalPages  int
}

// Snaps is a slice of Snap structs
var snaps = []Snap{
	{
		Name:     "Snap1",
		Address:  "192.168.137.173",
		Port:     "8081",
		Username: "hp",
	},
	{
		Name:     "Snap2",
		Address:  "192.168.75.28",
		Port:     "8081",
		Username: "elite",
	},
}

// User represents a user in the system
type User struct {
	ID        int
	Name      string
	Email     string
	CreatedAt string
}

// Product represents a product in the inventory
type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
	Quantity    int
	CreatedAt   string
}

// Order represents an order in the system
type Order struct {
	ID         int
	UserID     int
	TotalPrice float64
	OrderDate  string
	UserName   string
	Status     string
	ItemsCount int
}
type Customer struct {
	ID         int
	Name       string
	Email      string
	OrderCount int
	TotalSpent float64
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID          int
	OrderID     int
	ProductID   int
	Quantity    int
	Price       float64
	ProductName string
}

func NewPagination(currentPage, perPage, totalItems int) Pagination {
	totalPages := (totalItems + perPage - 1) / perPage
	return Pagination{
		CurrentPage: currentPage,
		PerPage:     perPage,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
	}
}

func (p Pagination) HasPrev() bool {
	return p.CurrentPage > 1
}

func (p Pagination) HasNext() bool {
	return p.CurrentPage < p.TotalPages
}

func (p Pagination) Pages() []int {
	pages := make([]int, 0, p.TotalPages)
	for i := 1; i <= p.TotalPages; i++ {
		pages = append(pages, i)
	}
	return pages
}

func CreateCustomer(name, email, password string) error {
	_, err := db.Exec("INSERT INTO users (name, email, password) VALUES (?, ?, ?)", name, email, password)
	return err
}

func GetCustomers() ([]User, error) {
	var users []User

	rows, err := db.Query("SELECT id, name, email, created_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func GetCustomerByID(id int) (User, error) {
	var user User

	row := db.QueryRow("SELECT id, name, email, created_at FROM users WHERE id = ?", id)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		return user, err
	}

	return user, nil
}

// UpdateCustomer updates an existing customer
func UpdateCustomer(id int, name, email string) error {
	_, err := db.Exec("UPDATE users SET name = ?, email = ? WHERE id = ?", name, email, id)
	return err
}

// DeleteCustomer removes a customer from the database
func DeleteCustomer(id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

// Product CRUD Operations

// CreateProduct adds a new product to the database
func CreateProduct(name, description string, price float64, quantity int) error {
	_, err := db.Exec("INSERT INTO products (name, description, price, quantity) VALUES (?, ?, ?, ?)",
		name, description, price, quantity)
	return err
}

// UpdateProduct updates an existing product
func UpdateProduct(id int, name, description string, price float64, quantity int) error {
	_, err := db.Exec("UPDATE products SET name = ?, description = ?, price = ?, quantity = ? WHERE id = ?",
		name, description, price, quantity, id)
	return err
}

// DeleteProduct removes a product from the database
func DeleteProduct(id int) error {
	_, err := db.Exec("DELETE FROM products WHERE id = ?", id)
	return err
}

// func DeleteProduct(id int) error {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return err
// 	}
// 	_, err = tx.Exec("DELETE FROM OrderItems WHERE product_id = ?", id)
// 	if err != nil {
// 		tx.Rollback()
// 		return err
// 	}
// 	_, err = tx.Exec("DELETE FROM Products WHERE id = ?", id)
// 	if err != nil {
// 		tx.Rollback()
// 		return err
// 	}
// 	return tx.Commit()
// }

func CreateOrder(userID int, totalPrice float64) (int64, error) {
	result, err := db.Exec("INSERT INTO orders (user_id, total_price) VALUES (?, ?)", userID, totalPrice)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateOrder updates an existing order
func UpdateOrder(id int, totalPrice float64) error {
	_, err := db.Exec("UPDATE orders SET total_price = ? WHERE id = ?", totalPrice, id)
	return err
}

func GetAllCustomersWithOrderStats() ([]Customer, error) {
	query := `
		SELECT 
			u.id,
			u.name,
			u.email,
			COUNT(o.id) AS order_count,
			COALESCE(SUM(o.total_price), 0) AS total_spent
		FROM 
			Users u
		LEFT JOIN 
			Orders o ON u.id = o.user_id
		GROUP BY 
			u.id, u.name, u.email
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []Customer

	for rows.Next() {
		var c Customer
		err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.OrderCount, &c.TotalSpent)
		if err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	fmt.Printf("Customers: %+v\n", customers)
	return customers, nil
}

func CreateOrderItem(orderID int64, productID, quantity int, price float64) error {
	_, err := db.Exec("INSERT INTO orderItems (order_id, product_id, quantity, price) VALUES (?, ?, ?, ?)",
		orderID, productID, quantity, price)
	return err
}
func UpdateOrderItem(id, quantity int, price float64) error {
	_, err := db.Exec("UPDATE order_items SET quantity = ?, price = ? WHERE id = ?",
		quantity, price, id)
	return err
}

// DeleteOrderItem removes an item from an order
func DeleteOrderItem(id int) error {
	_, err := db.Exec("DELETE FROM order_items WHERE id = ?", id)
	return err
}
func reportsHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Reports",
	}

	err := tmpl.ExecuteTemplate(w, "reports.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createOrderInDB(customerID, totalPrice string) (int, error) {
	var orderID int
	err := db.QueryRow("INSERT INTO orders (customer_id, total_price) VALUES (?, ?) RETURNING id", customerID, totalPrice).Scan(&orderID)
	if err != nil {
		return 0, err
	}
	return orderID, nil
}

func addProductToOrder(orderID, productID, quantity string) error {
	_, err := db.Exec("INSERT INTO order_items (order_id, product_id, quantity) VALUES (?, ?, ?)", orderID, productID, quantity)
	return err
}

func addTableFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.ParseFiles("templates/add_table.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

// Handler to process table creation form submission
func addTableSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Get form values
	tableName := r.FormValue("table_name")
	columns := r.Form["column_name"]
	dataTypes := r.Form["data_type"]

	// Here you would normally process the data and create the table in DB
	fmt.Printf("Creating table: %s\n", tableName)
	for i := range columns {
		fmt.Printf("Column %d: %s (%s)\n", i+1, columns[i], dataTypes[i])
	}

	// Redirect to success page or back to form
	http.Redirect(w, r, "/tables", http.StatusSeeOther)
}

// replicateToSlaves replicates database updates to all slave nodes
func replicateToSlaves(query string) {
	body := map[string]string{"query": query}
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error encoding JSON: %v", err)
		return
	}

	for _, snap := range snaps {
		url := fmt.Sprintf("http://%s:%s/replicate", snap.Address, snap.Port)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Failed to replicate to %s (%s): %v", snap.Name, snap.Address, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("%s returned non-OK status: %v", snap.Name, resp.Status)
		} else {
			log.Printf("Replication successful to %s (%s)", snap.Name, snap.Address)
		}
	}
}

// Database handlers

func addDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "add_database.html", map[string]interface{}{
		"Title": "Create New Database",
	})
}

// Database handlers for MySQL
func databasesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SHOW DATABASES;")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Skip system databases
		if name != "information_schema" && name != "mysql" && name != "performance_schema" && name != "sys" {
			databases = append(databases, name)
		}
	}

	data := struct {
		Title     string
		Databases []string
	}{
		Title:     "Databases",
		Databases: databases,
	}

	tmpl.ExecuteTemplate(w, "databases.html", data)
}

func createDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dbName := r.FormValue("name")
	if dbName == "" {
		http.Error(w, "Database name is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(fmt.Sprintf("CREATE DATABASE `%s`;", dbName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/databases", http.StatusSeeOther)
}

func deleteDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	dbName := pathParts[3]
	if dbName == "" {
		http.Error(w, "Database name is required", http.StatusBadRequest)
		return
	}

	// Can't delete the currently connected database
	if dbName == "ecommerce_db" {
		http.Error(w, "Cannot delete the currently connected database", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(fmt.Sprintf("DROP DATABASE `%s`;", dbName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/databases", http.StatusSeeOther)
}

func databaseAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// عرض نموذج إدخال بيانات الاعتماد
		dbName := r.URL.Query().Get("dbname")
		if dbName == "" {
			http.Error(w, "Database name is required", http.StatusBadRequest)
			return
		}

		tmpl.ExecuteTemplate(w, "database_auth.html", struct{ DBName string }{DBName: dbName})
		return
	}

	if r.Method == http.MethodPost {
		// معالجة بيانات الاعتماد المدخلة
		dbName := r.FormValue("dbname")
		username := r.FormValue("username")
		password := r.FormValue("password")

		if dbName == "" || username == "" || password == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		// تخزين بيانات الاعتماد مؤقتاً في الجلسة أو تمريرها كمعلمات
		// هنا سنمررها كمعلمات في الرابط
		redirectURL := fmt.Sprintf("/tables?db=%s&user=%s&password=%s",
			url.QueryEscape(dbName),
			url.QueryEscape(username),
			url.QueryEscape(password))

		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func databaseTablesHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	dbName := pathParts[2]
	if dbName == "" {
		http.Error(w, "Database name is required", http.StatusBadRequest)
		return
	}

	// Connect to the selected database
	selectedDB, err := sql.Open("mysql", fmt.Sprintf("user:password@tcp(127.0.0.1:3306)/%s", dbName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer selectedDB.Close()

	// Get tables in the selected database
	rows, err := selectedDB.Query("SHOW TABLES;")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tables = append(tables, tableName)
	}

	data := struct {
		Title    string
		Database string
		Tables   []string
	}{
		Title:    "Tables in " + dbName,
		Database: dbName,
		Tables:   tables,
	}

	tmpl.ExecuteTemplate(w, "database_tables.html", data)
}

// Table handlers

func tablesHandler(w http.ResponseWriter, r *http.Request) {
	// معالجة طلب GET لعرض الجداول
	if r.Method == http.MethodGet {
		dbName := r.URL.Query().Get("db")

		var rows *sql.Rows
		var err error

		if dbName != "" {
			// استخدام بيانات الاعتماد من المتغيرات البيئية أو القيم الافتراضية
			user := r.URL.Query().Get("user")
			password := r.URL.Query().Get("password")

			if user == "" {
				user = os.Getenv("DBUSER")
				if user == "" {
					user = "root"
				}
			}

			if password == "" {
				password = os.Getenv("DBPASS")
				if password == "" {
					password = "123456"
				}
			}

			// إنشاء اتصال جديد بقاعدة البيانات المحددة
			dbConn, err := sql.Open("mysql",
				fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", user, password, dbName))

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer dbConn.Close()

			rows, err = dbConn.Query("SHOW TABLES;")
		} else {
			// استخدام الاتصال الافتراضي
			rows, err = db.Query("SHOW TABLES;")
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var tables []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tables = append(tables, name)
		}

		data := struct {
			Title  string
			Tables []string
			DBName string
		}{
			Title:  "Tables",
			Tables: tables,
			DBName: dbName,
		}

		tmpl.ExecuteTemplate(w, "tables.html", data)
	}
}

func addTableHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "add_table.html", map[string]interface{}{
		"Title": "Create New Table",
	})
}

func createTableHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tableName := r.FormValue("name")
	if tableName == "" {
		http.Error(w, "Table name is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(fmt.Sprintf("CREATE TABLE %s (id SERIAL PRIMARY KEY);", tableName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/tables", http.StatusSeeOther)
}

func editTableHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	tableName := pathParts[3]
	if tableName == "" {
		http.Error(w, "Table name is required", http.StatusBadRequest)
		return
	}

	tmpl.ExecuteTemplate(w, "edit_table.html", map[string]interface{}{
		"Title":     "Edit Table",
		"TableName": tableName,
	})
}

func updateTableHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	oldName := pathParts[3]
	newName := r.FormValue("name")

	if oldName == "" || newName == "" {
		http.Error(w, "Table name is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", oldName, newName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/tables", http.StatusSeeOther)
}

func deleteTableHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	tableName := pathParts[3]
	if tableName == "" {
		http.Error(w, "Table name is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(fmt.Sprintf("DROP TABLE %s;", tableName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/tables", http.StatusSeeOther)
}

func addColumnHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	tableName := pathParts[3]
	if tableName == "" {
		http.Error(w, "Table name is required", http.StatusBadRequest)
		return
	}

	tmpl.ExecuteTemplate(w, "add_column.html", map[string]interface{}{
		"Title":     "Add Column",
		"TableName": tableName,
		"DataTypes": []string{"integer", "bigint", "serial", "bigserial", "text", "varchar", "char",
			"boolean", "date", "timestamp", "numeric", "real", "double precision", "json", "jsonb"},
	})
}

func createColumnHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	tableName := pathParts[3]
	columnName := r.FormValue("name")
	dataType := r.FormValue("type")
	isNullable := r.FormValue("nullable") == "true"
	defaultValue := r.FormValue("default")

	if tableName == "" || columnName == "" || dataType == "" {
		http.Error(w, "Table name, column name and type are required", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, dataType)
	if !isNullable {
		query += " NOT NULL"
	}
	if defaultValue != "" {
		query += " DEFAULT " + defaultValue
	}
	query += ";"

	_, err := db.Exec(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/table/columns/%s", tableName), http.StatusSeeOther)
}

func deleteColumnHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	tableName := pathParts[3]
	columnName := pathParts[4]

	if tableName == "" || columnName == "" {
		http.Error(w, "Table name and column name are required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", tableName, columnName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/table/columns/%s", tableName), http.StatusSeeOther)
}

// Column represents a database column
type Column struct {
	Name     string
	Type     string
	Nullable string
	Key      string
	Default  string
	Extra    string
}

// TableData represents the data for the table view
type TableData struct {
	Title     string
	DBName    string
	TableName string
	Columns   []Column
	Rows      [][]interface{}
}

func tableColumnsHandler(w http.ResponseWriter, r *http.Request) {
	tableName := r.URL.Path[len("/table/columns/"):]
	dbName := r.URL.Query().Get("db")

	// Connect to database (you'll need to implement your connection logic)
	db, err := sql.Open("mysql", "user:password@/"+dbName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get columns information
	columns, err := getColumns(db, tableName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get table data
	rows, err := getTableData(db, tableName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := TableData{
		Title:     "Table Columns",
		DBName:    dbName,
		TableName: tableName,
		Columns:   columns,
		Rows:      rows,
	}

	tmpl, err := template.ParseFiles("templates/table_columns.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

func getColumns(db *sql.DB, tableName string) ([]Column, error) {
	rows, err := db.Query(fmt.Sprintf("SHOW COLUMNS FROM %s", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var field, typ, null, key, extra string
		var def sql.NullString

		err := rows.Scan(&field, &typ, &null, &key, &def, &extra)
		if err != nil {
			return nil, err
		}

		columns = append(columns, Column{
			Name:     field,
			Type:     typ,
			Nullable: null,
			Key:      key,
			Default:  def.String,
			Extra:    extra,
		})
	}

	return columns, nil
}

func getTableData(db *sql.DB, tableName string) ([][]interface{}, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s LIMIT 100", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results [][]interface{}

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}

		results = append(results, values)
	}

	return results, nil
}
