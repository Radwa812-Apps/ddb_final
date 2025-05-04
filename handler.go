package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// *****************Customer***********
// *********************************************************************
func addCustomerSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		firstName := r.FormValue("firstName")
		email := r.FormValue("email")
		password := r.FormValue("password")

		var err error
		_, err = db.Exec("INSERT INTO users (name, email, password) VALUES (?, ?, ?)", firstName, email, password)
		if err != nil {
			log.Fatal(err)
		}
		// Replication to master
		query := fmt.Sprintf("INSERT INTO users (name, email, password) VALUES ('%s', '%s', '%s')",
			firstName, email, password)
		replicateToMaster(query)
		replicateToSlaves(query)
		// Redirect to customers page
		http.Redirect(w, r, "/customers", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/add-customer.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
func customersHandler(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}

	customers, err := GetAllCustomersWithOrderStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	total := len(customers)

	data := PageData{
		Title:          "Customers",
		CustomersStats: customers, // Pass the customers data
		Pagination:     NewPagination(page, 5, total),
	}

	err = tmpl.ExecuteTemplate(w, "customers.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func addCustomerHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Add Customer",
	}

	err := tmpl.ExecuteTemplate(w, "add-customer.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
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
	} else {
		_, err := db.Exec("UPDATE users SET name = ?, email = ?, password = ? WHERE id = ?", name, email, password, id)
		if err != nil {
			http.Error(w, "Error updating customer", http.StatusInternalServerError)
			return
		}
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

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	orders, err := GetOrders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title  string
		Orders []Order
	}{
		Title:  "Orders",
		Orders: orders,
	}

	err = tmpl.ExecuteTemplate(w, "orders.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/*func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	// Get customers and products for the form
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

*/

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

func productsHandler(w http.ResponseWriter, r *http.Request) {
	products, err := GetProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title    string
		Products []Product
	}{
		Title:    "Products",
		Products: products,
	}

	err = tmpl.ExecuteTemplate(w, "products.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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
