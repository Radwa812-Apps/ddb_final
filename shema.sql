

CREATE TABLE Users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT FALSE
);

CREATE TABLE Products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    quantity INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE Orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT,
    total_price DECIMAL(10, 2) NOT NULL,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'Processing',
    FOREIGN KEY (user_id) REFERENCES Users(id) ON DELETE CASCADE
);

CREATE TABLE OrderItems (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT,
    product_id INT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES Orders(id),
    FOREIGN KEY (product_id) REFERENCES Products(id) ON DELETE SET NULL
);



INSERT INTO Users (name, email, password)
VALUES 
    ('Alice Johnson', 'alice.johnson@example.com', 'alice123'),
    ('Bob Smith', 'bob.smith@example.com', 'bob456'),
    ('Charlie Brown', 'charlie.brown@example.com', 'charlie789'),
    ('David Williams', 'david.williams@example.com', 'david012'),
    ('Emma Davis', 'emma.davis@example.com', 'emma345');



INSERT INTO Products (name, description, price, quantity)
VALUES 
    ('Apple MacBook Pro 16-inch', 'Apple MacBook Pro with M1 Pro chip', 2499.00, 15),
    ('Samsung Galaxy S22', 'Samsung Galaxy S22 smartphone with 5G', 799.00, 30),
    ('Nike Air Max 270', 'Nike Air Max 270 shoes for men', 150.00, 50),
    ('Dell XPS 13', 'Dell XPS 13 laptop with Intel Core i7', 1499.00, 10),
    ('Sony WH-1000XM4', 'Sony noise-canceling wireless headphones', 349.00, 20);



INSERT INTO Orders (user_id, total_price)
VALUES 
    (1, 2499.00),  -- Alice Johnson
    (2, 799.00),   -- Bob Smith
    (3, 1499.00),  -- Charlie Brown
    (4, 1500.00),  -- David Williams
    (5, 349.00);   -- Emma Davis



INSERT INTO OrderItems (order_id, product_id, quantity, price)
VALUES
    (1, 1, 1, 2499.00),  -- Alice ordered 1 MacBook Pro
    (2, 2, 1, 799.00),   -- Bob ordered 1 Samsung Galaxy S22
    (3, 4, 1, 1499.00),  -- Charlie ordered 1 Dell XPS 13
    (4, 3, 2, 150.00),   -- David ordered 2 pairs of Nike Air Max 270
    (5, 5, 1, 349.00);   -- Emma ordered 1 Sony WH-1000XM4 headphones
