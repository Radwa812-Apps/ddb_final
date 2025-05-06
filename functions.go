package main

import (
	"database/sql"
	"fmt"
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


func GetPaginatedOrders(offset, limit int) ([]Order, error) {
    query := `
        SELECT o.id, o.user_id, o.total_price, o.order_date, u.name, 
               o.status, COUNT(oi.id) as items_count
        FROM orders o
        JOIN users u ON o.user_id = u.id
        LEFT JOIN orderitems oi ON o.id = oi.order_id
        GROUP BY o.id
        ORDER BY o.order_date DESC
        LIMIT ? OFFSET ?
    `
    
    rows, err := db.Query(query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orders []Order
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

func GetPaginatedProducts(offset, limit int) ([]Product, error) {
    query := `
        SELECT id, name, description, price, quantity, created_at 
        FROM products
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `
    
    rows, err := db.Query(query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []Product
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



func GetCustomerOrders(customerID int) ([]Order, error) {
    var orders []Order
    
    rows, err := db.Query(`
        SELECT o.id, o.total_price, o.order_date, o.status
        FROM orders o
        WHERE o.user_id = ?
        ORDER BY o.order_date DESC
    `, customerID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var order Order
        if err := rows.Scan(&order.ID, &order.TotalPrice, &order.OrderDate, &order.Status); err != nil {
            return nil, err
        }
        orders = append(orders, order)
    }

    return orders, nil
}

func GetPaginatedCustomersWithOrderStats(offset, limit int, orderBy string) ([]Customer, error) {
    query := fmt.Sprintf(`
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
        ORDER BY %s
        LIMIT ? OFFSET ?
    `, orderBy)

    rows, err := db.Query(query, limit, offset)
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

    return customers, nil
}







func GetAllCustomersWithOrderStats(page int, sort, search string) ([]Customer, error) {
    limit := 5
    offset := (page - 1) * limit
    
    baseQuery := `
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
        WHERE 1=1
    `
    
    // Add search condition if search term is provided
    if search != "" {
        baseQuery += " AND (u.name LIKE ? OR u.email LIKE ?)"
    }
    
    baseQuery += `
        GROUP BY 
            u.id, u.name, u.email
    `
    
    // Add sorting based on the sort parameter
    var orderBy string
    switch sort {
    case "oldest":
        orderBy = "u.created_at ASC"
    case "name_asc":
        orderBy = "u.name ASC"
    case "name_desc":
        orderBy = "u.name DESC"
    case "most_orders":
        orderBy = "order_count DESC"
    default: // "newest" (default)
        orderBy = "u.created_at DESC"
    }
    
    // Add LIMIT and OFFSET for pagination
    query := baseQuery + " ORDER BY " + orderBy + " LIMIT ? OFFSET ?"

    var rows *sql.Rows
    var err error
    
    if search != "" {
        searchTerm := "%" + search + "%"
        rows, err = db.Query(query, searchTerm, searchTerm, limit, offset)
    } else {
        rows, err = db.Query(query, limit, offset)
    }
    
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

    return customers, nil
}