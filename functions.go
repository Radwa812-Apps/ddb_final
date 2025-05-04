package main

import (
	"strconv"
	"strings"
)

func GetProducts() ([]Product, error) {
	var products []Product

	rows, err := db.Query("SELECT id, name, description, price, quantity, created_at FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Quantity, &product.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}
func GetProductByID(id int) (Product, error) {
	var product Product

	row := db.QueryRow("SELECT id, name, description, price, quantity, created_at FROM products WHERE id = ?", id)
	err := row.Scan(&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Quantity, &product.CreatedAt)
	if err != nil {
		return product, err
	}

	return product, nil
}

func GetOrderItems(orderID int) ([]OrderItem, error) {
	var items []OrderItem

	rows, err := db.Query(`
    SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, p.name 
    FROM orderitems oi  // Changed from order_items
    JOIN products p ON oi.product_id = p.id
    WHERE oi.order_id = ?
`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity,
			&item.Price, &item.ProductName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
func toInt(s string) int {
	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0 // or handle it differently depending on your needs
	}
	return i
}

func toFloat(s string) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0.0 // or handle it differently
	}
	return f
}


func GetOrders() ([]Order, error) {
    var orders []Order
    rows, err := db.Query(`
        SELECT o.id, o.user_id, o.total_price, o.order_date, u.name, 
               o.status, COUNT(oi.id) as items_count
        FROM orders o
        JOIN users u ON o.user_id = u.id
        LEFT JOIN orderitems oi ON o.id = oi.order_id
        GROUP BY o.id
        ORDER BY o.order_date DESC
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var order Order
        if err := rows.Scan(&order.ID, &order.UserID, &order.TotalPrice, 
            &order.OrderDate, &order.UserName, &order.Status, &order.ItemsCount); err != nil {
            return nil, err
        }
        orders = append(orders, order)
    }

    return orders, nil
}

func GetOrderByID(id int) (Order, error) {
    var order Order

    row := db.QueryRow(`
        SELECT o.id, o.user_id, o.total_price, o.order_date, u.name, 
               o.status, COUNT(oi.id) as items_count
        FROM orders o
        JOIN users u ON o.user_id = u.id
        LEFT JOIN order_items oi ON o.id = oi.order_id
        WHERE o.id = ?
        GROUP BY o.id
    `, id)
    err := row.Scan(&order.ID, &order.UserID, &order.TotalPrice, 
        &order.OrderDate, &order.UserName, &order.Status, &order.ItemsCount)
    if err != nil {
        return order, err
    }

    return order, nil
}