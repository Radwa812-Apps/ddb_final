# 🛍️ E-Commerce Database Management System with Replication

## 📌 Overview

A distributed e-commerce database system with master-slave replication architecture, built with Go and MySQL. Provides complete CRUD operations for customers, products, and orders through a web interface with automatic synchronization across nodes.

## ✨ Key Features

### 🔄 Replication System

- **Master-Slave Architecture** with automatic synchronization
- **Bi-directional replication** capability
- **Conflict resolution** for data consistency
- **Real-time updates** across all nodes

### 🛒 E-Commerce Modules

| Module      | Operations                    |
| ----------- | ----------------------------- |
| Customers   | Create, Read, Update, Delete  |
| Products    | Inventory management, Pricing |
| Orders      | Processing, Cancellation      |
| Order Items | Detailed product tracking     |

### ⚙️ Technical Features

- **Pagination** for large datasets
- **Advanced search** with filters
- **Sorting** by multiple criteria
- **Database administration** tools
- **Responsive web interface**

## 🛠️ Installation Guide

### Prerequisites

- Go 1.16+
- MySQL 5.7+
- Git

### Setup Instructions

```bash
# Clone repository
git clone https://github.com/Radwa812-Apps/ddb_final/tree/final_slave
cd ecommerce-dbms

# Set environment variables
export DBUSER=your_db_username
export DBPASS=your_db_password

# Initialize database
mysql -u root -p -e "CREATE DATABASE ecommerce_db"
mysql -u root -p ecommerce_db < schema.sql

# Build and run
go run .

# Start master node
./ecommerce-dbms

# In another terminal (slave node)
DBPORT=8081 ./ecommerce-dbms
```

## ⚙️ Configuration

### Master Node (crud_operation.go)

```go
var snaps = []Snap{
    {
        Name:     "Alyaa-Slave", 
        Address:  "192.168.39.155",
        Port:     "8081",
        Username: "elite",
        Database: "ecommerce_db",
    },
    // Add more slave nodes as needed
}
```

### Environment Variables

| Variable | Description       | Example      |
| -------- | ----------------- | ------------ |
| DBUSER   | Database username | elite        |
| DBPASS   | Database password | securepass   |
| DBPORT   | Web server port   | 8080 or 8081 |

## 🌐 Web Interface Endpoints

### Core Routes

| Endpoint       | Description         | Methods   |
| -------------- | ------------------- | --------- |
| `/customers` | Customer management | GET, POST |
| `/products`  | Product inventory   | GET, POST |
| `/orders`    | Order processing    | GET, POST |
| `/dashboard` | System overview     | GET       |

### API Endpoints

| Endpoint        | Description                 |
| --------------- | --------------------------- |
| `/replicate`  | Replication endpoint (POST) |
| `/api/orders` | JSON order data             |

## 🔄 Replication Flow

```mermaid
sequenceDiagram
    participant Client
    participant Slave1
    participant Slave2
    participant Master

    Client->>Slave1: INSERT INTO products...
    Slave1->>Slave1DB: Execute Query 
    Slave1->>Slave2: POST /replicate
    Slave2->>Slave2DB: Execute Query
    Slave1->>Master: POST /replicate 
    Master->>MasterDB: Execute Query
  
```

## 📂 Project Structure

```
ecommerce-dbms/
├── crud_operation.go    # Core database operations
├── function.go          # Business logic
├── handler.go           # HTTP controllers  
├── schema.sql           # database schema
├── slave.go             # Slave node entry
├── templates/           # HTML views
│   ├── base.html
│   ├── customers.html
│   └── ...
└── README.md
```

## 🐛 Troubleshooting Guide

| Symptom                    | Solution                   |
| -------------------------- | -------------------------- |
| Replication failures       | Check network connectivity |
| Database connection issues | Verify credentials in cfg  |
| Template rendering errors  | Validate template syntax   |
| Data inconsistency         | Check replication logs     |

## 📜 License

This project is licensed under the  **MIT License** .

## 📧 Contact

📧 Email: Shaima.AbdulRahim829@compit.aun.edu.eg
🔗 GitHub: https://github.com/Radwa812-Apps

---

```bash
# Quick Start
make deploy && ./ecommerce-dbms
```

> **Note**: Ensure all nodes can communicate on the specified ports (8080 for master, 8081 for slave by default)
