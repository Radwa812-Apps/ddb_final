package main

import (
	"encoding/json"
	"fmt"

	"log"
	"net/http"
	"strconv"
	"strings"
)

// *****************Customer***********
// *********************************************************************

func deleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/customer/delete/")
	if id == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Disable foreign key checks
	_, err = tx.Exec("SET FOREIGN_KEY_CHECKS=0")
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete user
	_, err = tx.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	query := fmt.Sprintf("DELETE FROM users WHERE id = '%s'", id)
	replicateToSlaves(query, "root", "rootroot")
	// Re-enable checks and commit
	_, err = tx.Exec("SET FOREIGN_KEY_CHECKS=1")
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/customers", http.StatusSeeOther)
}
func updateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if password == "" {
		_, err := db.Exec("UPDATE users SET name = ?, email = ? WHERE id = ?", name, email, id)
		if err != nil {
			http.Error(w, "Error updating customer", http.StatusInternalServerError)
			return
		}
		query := fmt.Sprintf("UPDATE users SET name = '%s', email = '%s' WHERE id = %s", name, email, id)
		fmt.Println(query)
		replicateToSlaves(query, "root", "rootroot")

	} else {
		_, err := db.Exec("UPDATE users SET name = ?, email = ?, password = ? WHERE id = ?", name, email, password, id)
		if err != nil {
			http.Error(w, "Error updating customer", http.StatusInternalServerError)
			return
		}
		query := fmt.Sprintf("UPDATE users SET name = '%s', email = '%s', password = '%s' WHERE id = %s", name, email, password, id)
		fmt.Println(query)
		replicateToSlaves(query, "root", "rootroot")
	}

	http.Redirect(w, r, "/customers", http.StatusSeeOther)
}

func editCustomerHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/customer/edit/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	customer, err := GetCustomerByID(id)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	data := struct {
		Title    string
		Customer User
	}{
		Title:    "Edit Customer",
		Customer: customer,
	}

	tmpl.ExecuteTemplate(w, "edit-customer.html", data)
}

//****************************Order**************
//********************************************************

func cancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/order/cancel/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Update order status to cancelled
	_, err = db.Exec("UPDATE orders SET status = 'Cancelled' WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Error cancelling order", http.StatusInternalServerError)
		return
	}
	query := fmt.Sprintf("UPDATE orders SET status = 'Cancelled' WHERE id = %d", id)
	replicateToSlaves(query, "root", "rootroot")
	http.Redirect(w, r, "/orders", http.StatusSeeOther)
}

/**/

func updateOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	totalPrice := r.FormValue("total_price")

	err := UpdateOrder(toInt(id), toFloat(totalPrice))
	if err != nil {
		http.Error(w, "Error updating order", http.StatusInternalServerError)
		return
	}
	query := fmt.Sprintf("UPDATE orders SET total_price = %s WHERE id = %s", totalPrice, id)
	replicateToSlaves(query, "root", "rootroot")
	w.Write([]byte("Order updated successfully"))
}

func viewOrderHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	order, err := GetOrderByID(toInt(id))
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(order)
}

//***************Product handlers***********

func createProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// Get form values
	name := r.FormValue("name")
	description := r.FormValue("description")
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		http.Error(w, "Invalid price format", http.StatusBadRequest)
		return
	}
	quantity, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		http.Error(w, "Invalid quantity format", http.StatusBadRequest)
		return
	}

	// Create product in database
	err = CreateProduct(name, description, price, quantity)
	if err != nil {
		http.Error(w, "Error creating product", http.StatusInternalServerError)
		return
	}
	// Replication
	query := fmt.Sprintf("INSERT INTO products (name, description, price, quantity) VALUES ('%s', '%s', %f, %d)",
		name, description, price, quantity)
	replicateToSlaves(query, "root", "rootroot")
	// Redirect to products page
	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

func deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/product/delete/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = DeleteProduct(id)
	if err != nil {
		http.Error(w, "Error deleting product", http.StatusInternalServerError)
		return
	}
	query := fmt.Sprintf("DELETE FROM products WHERE id = %d", id)
	replicateToSlaves(query, "root", "rootroot")
	http.Redirect(w, r, "/products", http.StatusSeeOther)
}
func updateProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		http.Error(w, "Invalid price", http.StatusBadRequest)
		return
	}
	quantity, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	err = UpdateProduct(id, name, description, price, quantity)
	if err != nil {
		http.Error(w, "Error updating product", http.StatusInternalServerError)
		return
	}
	query := fmt.Sprintf("UPDATE products SET name = '%s', description = '%s', price = %f, quantity = %d WHERE id = %d",
		name, description, price, quantity, id)
	replicateToSlaves(query, "root", "rootroot")
	http.Redirect(w, r, "/products", http.StatusSeeOther)
}
func editProductHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/product/edit/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := GetProductByID(id)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	data := struct {
		Title   string
		Product Product
	}{
		Title:   "Edit Product",
		Product: product,
	}

	tmpl.ExecuteTemplate(w, "edit-product.html", data)
}
func addProductHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
	}{
		Title: "Add Product",
	}
	tmpl.ExecuteTemplate(w, "add-product.html", data)
}
func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Parse form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form data", http.StatusBadRequest)
			return
		}

		// Get form values
		userID, err := strconv.Atoi(r.FormValue("user_id"))
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		totalPrice, err := strconv.ParseFloat(r.FormValue("total_price"), 64)
		if err != nil {
			http.Error(w, "Invalid total price", http.StatusBadRequest)
			return
		}

		productIDs := r.Form["product_id[]"]
		quantities := r.Form["quantity[]"]
		prices := r.Form["price[]"]

		// Validate products
		if len(productIDs) == 0 || len(productIDs) != len(quantities) || len(productIDs) != len(prices) {
			http.Error(w, "Invalid product data", http.StatusBadRequest)
			return
		}

		// Create order
		orderID, err := CreateOrder(userID, totalPrice)
		if err != nil {
			http.Error(w, "Error creating order: "+err.Error(), http.StatusInternalServerError)
			return
		}
		query := fmt.Sprintf("INSERT INTO orders (user_id, total_price) VALUES (%d, %f)", userID, totalPrice)
		replicateToSlaves(query, "root", "rootroot")
		// Add order items
		for i, productIDStr := range productIDs {
			productID, err := strconv.Atoi(productIDStr)
			if err != nil {
				http.Error(w, "Invalid product ID", http.StatusBadRequest)
				return
			}

			quantity, err := strconv.Atoi(quantities[i])
			if err != nil {
				http.Error(w, "Invalid quantity", http.StatusBadRequest)
				return
			}

			price, err := strconv.ParseFloat(prices[i], 64)
			if err != nil {
				http.Error(w, "Invalid price", http.StatusBadRequest)
				return
			}

			err = CreateOrderItem(orderID, productID, quantity, price)
			if err != nil {
				http.Error(w, "Error adding order item: "+err.Error(), http.StatusInternalServerError)
				return
			}
			query := fmt.Sprintf("INSERT INTO orderItems (order_id, product_id, quantity, price) VALUES (%d, %d, %d, %f)",
				orderID, productID, quantity, price)
			replicateToSlaves(query, "root", "rootroot")
		}

		http.Redirect(w, r, "/orders", http.StatusSeeOther)
		return
	}

	// If GET request, render the order creation form
	customers, err := GetCustomers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	products, err := GetProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title     string
		Customers []User
		Products  []Product
	}{
		Title:     "Create Order",
		Customers: customers,
		Products:  products,
	}

	err = tmpl.ExecuteTemplate(w, "add-order.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}

	limit := 5
	offset := (page - 1) * limit

	products, err := GetPaginatedProducts(offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get total count for pagination
	var total int
	err = db.QueryRow("SELECT COUNT(*) FROM products").Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title      string
		Products   []Product
		Pagination Pagination
	}{
		Title:      "Products",
		Products:   products,
		Pagination: NewPagination(page, limit, total),
	}

	err = tmpl.ExecuteTemplate(w, "products.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}

	limit := 5
	offset := (page - 1) * limit

	orders, err := GetPaginatedOrders(offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get total count for pagination
	var total int
	err = db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title      string
		Orders     []Order
		Pagination Pagination
	}{
		Title:      "Orders",
		Orders:     orders,
		Pagination: NewPagination(page, limit, total),
	}

	err = tmpl.ExecuteTemplate(w, "orders.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewCustomerHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/customer/view/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	// Get customer details
	customer, err := GetCustomerByID(id)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	// Get customer orders
	orders, err := GetCustomerOrders(id)
	if err != nil {
		http.Error(w, "Error fetching orders", http.StatusInternalServerError)
		return
	}

	data := struct {
		Customer User
		Orders   []Order
	}{
		Customer: customer,
		Orders:   orders,
	}

	tmpl.ExecuteTemplate(w, "view-customer.html", data)
}

func addCustomerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Handle POST request (form submission)
		firstName := r.FormValue("firstName")
		email := r.FormValue("email")
		password := r.FormValue("password")

		err := CreateCustomer(firstName, email, password)
		if err != nil {
			log.Printf("Error creating customer: %v", err)
			http.Error(w, "Error creating customer", http.StatusInternalServerError)
			return
		}

		// Replication
		query := fmt.Sprintf("INSERT INTO users (name, email, password) VALUES ('%s', '%s', '%s')",
			firstName, email, password)

		replicateToSlaves(query, "root", "rootroot")

		http.Redirect(w, r, "/customers", http.StatusSeeOther)
		return
	}

	// Handle GET request (show form)
	data := PageData{
		Title: "Add Customer",
	}
	err := tmpl.ExecuteTemplate(w, "add-customer.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func customersHandler(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}

	// Get sort parameter
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "newest" // default sort
	}

	// Get search query parameter
	search := r.URL.Query().Get("search")

	// Get customers with search, sort, and pagination
	customers, err := GetAllCustomersWithOrderStats(page, sort, search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get total count of customers (with search filter if applicable)
	var total int
	var countQuery string
	var countArgs []interface{}

	if search != "" {
		countQuery = "SELECT COUNT(*) FROM users WHERE name LIKE ? OR email LIKE ?"
		searchTerm := "%" + search + "%"
		countArgs = []interface{}{searchTerm, searchTerm}
	} else {
		countQuery = "SELECT COUNT(*) FROM users"
	}

	err = db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare data for template
	data := PageData{
		Title:          "Customers",
		CustomersStats: customers,
		Pagination:     NewPagination(page, 5, total),
		Sort:           sort,
		SearchQuery:    search, // Pass search query to template
	}

	// Execute template
	err = tmpl.ExecuteTemplate(w, "customers.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
