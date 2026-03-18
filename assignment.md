---
---
---
Name: "Suraj Gorai"
Github: "github.com/surajgoraicse"
Email : "surajgoraicse@gmail.com"
Linkedin : "linkedin.com/in/surajgoraicse"
Ph no : 9110099518
date: 2026-03-18

---



# SQL Ques : 1

```sql
CREATE TABLE Products (
    ProductID INT PRIMARY KEY,
    ProductName VARCHAR(255) NOT NULL,
    Price DECIMAL(10, 2) NOT NULL,
    StockQuantity INT NOT NULL);

CREATE TABLE Orders (
    OrderID INT PRIMARY KEY,
    ProductID INT,
    OrderQuantity INT NOT NULL,
    OrderDate DATETIME DEFAULT CURRENT_TIMESTAMP,
    OrderStatus VARCHAR(50) DEFAULT 'Pending',
    FOREIGN KEY (ProductID) REFERENCES Products(ProductID) );

```

- **Design a stored procedure named `PlaceOrder` that handles new customer orders. This procedure should:**
  Accept `p_ProductID` and `p_OrderQuantity` as input parameters.
- **Check Stock:** Before placing an order, verify if there is sufficient `StockQuantity` for the given `p_ProductID` in the Products table.
- **Place Order (if sufficient stock):** If stock is sufficient, insert a new record into the Orders table with the provided `p_ProductID` and `p_OrderQuantity`.
  update the `StockQuantity` in the Products table by decreasing it by the p_OrderQuantity. Set the OrderStatus to 'Confirmed'.
- **Handle Insufficient Stock:** If stock is insufficient, do not insert into the Orders table.
  Instead, you should conceptually indicate that the order could not be placed (e.g., via an output parameter, or by simply not inserting the order and leaving the OrderStatus as 'Failed' if you were to insert a placeholder, but for this problem, we'll assume no insertion if stock is low).
- **Output:** The procedure should ideally provide a way to indicate success or failure (e.g., using an OUT parameter).

---

### solution

- Here I am assuming that we are using `pgx` driver for postgres db and `pgxpool` for safe connection pool.

```go
// type definition for Order object
type Order struct {
    OrderID int  `db:"OrderID"`
    ProductID int `db:"ProductID"`
    OrderQuantity int `db:"OrderQuantity"`
    OrderDate time.Time `db:"OrderDate"`
    OrderStatus string `db:"OrderStatus"`
}

// custom error for insufficient stock
var (
    InsufficientStockError = error.New("insufficient stock")
)

```

```go
func PlaceOrder(ctx context.Context, pool *pgxpool.Pool, productID int, orderQuantity int) (Order, error) {

    var order Order

    // Start a transaction
    // pool is *pgxpool.Pool
    tx, err := pool.Begin(ctx)
    if err != nil {
        return order, err
    }
    defer tx.Rollback(ctx)

    // Check stock
    var stockQuantity int
    err = tx.QueryRow(ctx, "SELECT StockQuantity FROM Products WHERE ProductID = $1", productID).Scan(&stockQuantity)
    if err != nil {
        return order, err
    }

    if stockQuantity < orderQuantity {
        return order, InsufficientStockError
    }

    // Generate order ID
    var orderID int
    err = tx.QueryRow(ctx, "SELECT COALESCE(MAX(OrderID), 0) + 1 FROM Orders").Scan(&orderID)
    if err != nil {
        return order, err
    }

    // Place order
    err = tx.QueryRow(ctx, `
        INSERT INTO Orders (OrderID, ProductID, OrderQuantity) 
        VALUES ($1, $2, $3) 
        RETURNING OrderID, ProductID, OrderQuantity, OrderDate, OrderStatus
    `, orderID, productID, orderQuantity).Scan(&order.OrderID, &order.ProductID, &order.OrderQuantity, &order.OrderDate, &order.OrderStatus)
    if err != nil {
        return order, err
    }

    // Update stock
    _, err = tx.Exec(ctx, "UPDATE Products SET StockQuantity = StockQuantity - $1 WHERE ProductID = $2", orderQuantity, productID)
    if err != nil {
        return order, err
    }

    // Set Order status to confirmed
    err = tx.QueryRow(ctx, "UPDATE Orders SET OrderStatus = 'Confirmed' WHERE OrderID = $1 RETURNING OrderStatus", orderID).Scan(&order.OrderStatus)
    if err != nil {
        return order, err
    }

    // Commit transaction
    return order, tx.Commit(ctx)
}

```

### Note to the Evaluator :

1. There is a critical problem in the DB design. The ID are of type `int`. This means that we have to manually generate the ID for each new record.
    - Here I am using `SELECT COALESCE(MAX(OrderID), 0) + 1 FROM Orders` to generate the ID, but this is not a good practice as it can lead to race conditions.
    - Even if we use `SERIAL` type or INT with `AUTO_INCREMENT` will fail in distributed environments.
    - Solution : Replace the `int` type with `UUID` type for the ID columns and use `uuidv7()` inbuild postgres18 function to generate the ID.
        - This guarantee globally uniqueness across all systems.

2. According to the problem statement we have to first set the `OrderStatus` to "pending" and after updating the stock we have to set the `OrderStatus` to "confirmed". This is a double write on `Orders`. If the Order is successfully placed with can set it to `Confirmed` directly. If a transaction fails, we will send a failed response to the client.
3. We can add a Postgres enum type for order status for consistency : 
example : 
```sql
CREATE TYPE order_status AS ENUM (
    'pending',
    'processing',
    'shipped',
    'delivered',
    'cancelled'
);
```


# SQL 2 : 

```sql

CREATE TABLE Departments (
    DepartmentID INT PRIMARY KEY,
    DepartmentName VARCHAR(100) NOT NULL);
    
CREATE TABLE Employees (
    EmployeeID INT PRIMARY KEY,
    EmployeeName VARCHAR(100) NOT NULL,
    DepartmentID INT,
    HireDate DATE NOT NULL,
    FOREIGN KEY (DepartmentID) REFERENCES Departments(DepartmentID));
    
CREATE TABLE Salaries (
    SalaryID INT PRIMARY KEY,
    EmployeeID INT,
    SalaryAmount DECIMAL(10, 2) NOT NULL,
    EffectiveDate DATE NOT NULL,
    FOREIGN KEY (EmployeeID) REFERENCES Employees(EmployeeID));
    
```

Write a SQL query to retrieve the following information for each department:
- DepartmentName
- Number of Employees in that department

Average Latest Salary of employees in that department (based on the most recent   EffectiveDate per employee)

### solution 

- here salary table has multiple rows per employee, but we only want the latest one.

```sql
WITH LatestSalary AS (
    SELECT 
        EmployeeID,
        SalaryAmount,
        ROW_NUMBER() OVER (
            PARTITION BY EmployeeID 
            ORDER BY EffectiveDate DESC
        ) AS rn
    FROM Salaries
)

SELECT 
    d.DepartmentName,
    COUNT(e.EmployeeID) AS EmployeeCount,
    AVG(ls.SalaryAmount) AS AvgLatestSalary
FROM Departments d
LEFT JOIN Employees e 
    ON d.DepartmentID = e.DepartmentID
LEFT JOIN LatestSalary ls 
    ON e.EmployeeID = ls.EmployeeID 
    AND ls.rn = 1
GROUP BY d.DepartmentName;
```


# Coding Ques : 

## Question 1: 

Write a function that takes any character as input and returns a compressed version of the string.
Input:
Example: "aaabbbcccd"
Output:
Example: "a3b3c3d1"

### Solution : 

also written test cases : 
run tests using : `go test ./...`
[here is the main.go file](./main.go)



```go
func Compress(s string) string {
	if len(s) == 0 {
		return ""
	}
	var (
		count   = 0
		pointer = rune(s[0])
	)
	var result strings.Builder  // for better performance

	for _, v := range s {
		if v == pointer {
			count++
		} else {
			result.WriteRune(pointer)
			result.WriteString(strconv.Itoa(count))
			pointer = v
			count = 1
		}
	}
	result.WriteRune(pointer)
	result.WriteString(strconv.Itoa(count))

	return result.String()
}
```


## Question 2 : 

Given 8 digits, return the latest valid date and time that can be formed in the format

"YYYY-MM-DD HH:MM"  
Input:[2, 0, 2, 3, 1, 1, 3, 0]  
Output: "2023-11-30 13:00"


### Note to evaluator : 

I believe there is some mistake with the question :
- There are only 8 digits in the input and the output expects 12 digits. 
	- generating `"YYYY-MM-DD HH:MM"` require 12 digits.
- If I assume that we can reuse the digits then and generate the most latest date and time then the given example is possibly wrong : 

```go
"YYYY-MM-DD HH:MM"  
Input:[2, 0, 2, 3, 1, 1, 3, 0]  
Output: "2023-11-30 13:00"

since we can reuse digits then : "2023-12-30 23:33" will me more latest and also valid.  
```

or if we assume that we cannot reuse the digits then, again the example use case will be wrong as it has use `3` more than its occurrence : 

```go
"YYYY-MM-DD HH:MM"  
Input:[2, 0, 2, 3, 1, 1, 3, 0]  
Output: "2023-11-30 13:00"

here 3 is used thrice, while they are only twice in the array.
```




