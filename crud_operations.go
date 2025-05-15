package main

import (
	_ "github.com/go-sql-driver/mysql"

	_ "github.com/lib/pq"
)

// Snaps represents the list of slave nodes
type Snap struct {
	Name     string
	Address  string
	Port     string
	Username string
	Database string
}

// Snaps is a slice of Snap structs
var snaps = []Snap{

	{
		Name:     "Alyaa",
		Address:  "192.168.39.155", //Alyaa
		Port:     "8081",
		Username: "elite",
		Database: "ecommerce_db",
	},
	
}

type User struct {
	ID        int
	Name      string
	Email     string
	CreatedAt string
}
type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
	Quantity    int
	CreatedAt   string
}

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

type OrderItem struct {
	ID          int
	OrderID     int
	ProductID   int
	Quantity    int
	Price       float64
	ProductName string
}
type Pagination struct {
	CurrentPage int
	PerPage     int
	TotalItems  int
	TotalPages  int
}

// *************Customer********************/
func UpdateCustomer(id int, name, email string) error {
	_, err := db.Exec("UPDATE users SET name = ?, email = ? WHERE id = ?", name, email, id)
	return err
}
func DeleteCustomer(id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}
func CreateCustomer(name, email, password string) error {
	_, err := db.Exec("INSERT INTO users (name, email, password) VALUES (?, ?, ?)", name, email, password)
	return err
}

/**************Product*********************/

func CreateProduct(name, description string, price float64, quantity int) error {
	_, err := db.Exec("INSERT INTO products (name, description, price, quantity) VALUES (?, ?, ?, ?)",
		name, description, price, quantity)
	return err
}

func UpdateProduct(id int, name, description string, price float64, quantity int) error {
	_, err := db.Exec("UPDATE products SET name = ?, description = ?, price = ?, quantity = ? WHERE id = ?",
		name, description, price, quantity, id)
	return err
}
func DeleteProduct(id int) error {
	_, err := db.Exec("DELETE FROM products WHERE id = ?", id)
	return err
}

/**********Order********/
func CreateOrder(userID int, totalPrice float64) (int64, error) {
	result, err := db.Exec("INSERT INTO orders (user_id, total_price) VALUES (?, ?)", userID, totalPrice)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
func UpdateOrder(id int, totalPrice float64) error {
	_, err := db.Exec("UPDATE orders SET total_price = ? WHERE id = ?", totalPrice, id)
	return err
}

/**********orderItem Bridge**********/

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
func DeleteOrderItem(id int) error {
	_, err := db.Exec("DELETE FROM order_items WHERE id = ?", id)
	return err
}
