package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Global database connection shared across the application
var db *sql.DB

// Simple HTTP client for external API calls
var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

// InitializeDatabase sets up the database connection and ensures schema exists
func InitializeDatabase() error {
	var err error

	// Open database connection
	db, err = sql.Open("sqlite3", "bookstore.db")
	if err != nil {
		return err
	}

	// Configure connection pool for optimal concurrent performance
	db.SetMaxOpenConns(25)                 // Maximum total connections
	db.SetMaxIdleConns(25)                 // Keep connections alive for reuse
	db.SetConnMaxLifetime(5 * time.Minute) // Refresh connections periodically

	// Smart initialization - only setup if needed
	return initializeDatabaseIfNeeded()
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// initializeDatabaseIfNeeded checks if database is already set up before running setup
func initializeDatabaseIfNeeded() error {
	// Test if database is already initialized by checking if books table exists and has data
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)

	// If query succeeded and we have data, database is already initialized
	if err == nil && count > 0 {
		log.Printf("Database already initialized with %d books, skipping setup", count)
		return nil
	}

	// If we get here, either:
	// 1. Table doesn't exist (query failed)
	// 2. Table exists but is empty (count = 0)
	// Either way, we need to run setup

	log.Println("Initializing database schema and data...")

	if err := createSchema(); err != nil {
		return err
	}

	if err := populateInitialData(); err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

// createSchema creates all necessary database tables
func createSchema() error {
	// Create books table for basic metadata
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS books (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			author TEXT NOT NULL,
			isbn TEXT UNIQUE,
			publish_date DATE,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create pricing table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS pricing (
			book_id TEXT PRIMARY KEY,
			price DECIMAL(10,2) NOT NULL,
			currency TEXT DEFAULT 'USD',
			discount DECIMAL(3,2) DEFAULT 0.0,
			sale_price DECIMAL(10,2),
			promotion TEXT,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (book_id) REFERENCES books(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create inventory table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS inventory (
			book_id TEXT PRIMARY KEY,
			in_stock BOOLEAN DEFAULT true,
			quantity INTEGER DEFAULT 0,
			warehouse TEXT,
			shipping_time TEXT,
			last_restocked TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (book_id) REFERENCES books(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create reviews table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS reviews (
			book_id TEXT PRIMARY KEY,
			average_rating DECIMAL(2,1),
			total_reviews INTEGER DEFAULT 0,
			recent_review TEXT,
			five_star INTEGER DEFAULT 0,
			four_star INTEGER DEFAULT 0,
			three_star INTEGER DEFAULT 0,
			two_star INTEGER DEFAULT 0,
			one_star INTEGER DEFAULT 0,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (book_id) REFERENCES books(id)
		)
	`)

	return err
}

// populateInitialData inserts sample data into all tables
func populateInitialData() error {
	// Insert book metadata
	books := []map[string]interface{}{
		{"id": "1", "title": "The Go Programming Language", "author": "Alan Donovan", "isbn": "978-0134190440", "publish_date": "2015-11-16", "description": "The authoritative resource to writing clear and idiomatic Go"},
		{"id": "2", "title": "Clean Code", "author": "Robert Martin", "isbn": "978-0132350884", "publish_date": "2008-08-11", "description": "A handbook of agile software craftsmanship"},
		{"id": "3", "title": "System Design Interview", "author": "Alex Xu", "isbn": "978-1736049112", "publish_date": "2020-06-04", "description": "An insider's guide to system design interviews"},
		{"id": "4", "title": "Dopamine Nation", "author": "Anna Lembke", "isbn": "978-1524746728", "publish_date": "2021-08-24", "description": "Finding balance in the age of indulgence"},
	}

	for _, book := range books {
		_, err := db.Exec(`
			INSERT OR IGNORE INTO books (id, title, author, isbn, publish_date, description) 
			VALUES (?, ?, ?, ?, ?, ?)
		`, book["id"], book["title"], book["author"], book["isbn"], book["publish_date"], book["description"])
		if err != nil {
			return err
		}
	}

	// Insert pricing data
	pricing := []map[string]interface{}{
		{"book_id": "1", "price": 39.99, "discount": 0.10, "sale_price": 35.99, "promotion": "Holiday Sale"},
		{"book_id": "2", "price": 32.50, "discount": 0.05, "sale_price": 30.88, "promotion": "Member Discount"},
		{"book_id": "3", "price": 28.95, "discount": 0.00, "sale_price": 28.95, "promotion": ""},
		{"book_id": "4", "price": 20.00, "discount": 0.15, "sale_price": 17.00, "promotion": "Limited Time"},
	}

	for _, p := range pricing {
		_, err := db.Exec(`
			INSERT OR IGNORE INTO pricing (book_id, price, discount, sale_price, promotion) 
			VALUES (?, ?, ?, ?, ?)
		`, p["book_id"], p["price"], p["discount"], p["sale_price"], p["promotion"])
		if err != nil {
			return err
		}
	}

	// Insert inventory data
	inventory := []map[string]interface{}{
		{"book_id": "1", "in_stock": true, "quantity": 42, "warehouse": "East Coast DC", "shipping_time": "2-3 business days"},
		{"book_id": "2", "in_stock": true, "quantity": 38, "warehouse": "Central DC", "shipping_time": "1-2 business days"},
		{"book_id": "3", "in_stock": true, "quantity": 15, "warehouse": "West Coast DC", "shipping_time": "3-4 business days"},
		{"book_id": "4", "in_stock": false, "quantity": 0, "warehouse": "Back Order", "shipping_time": "2-3 weeks"},
	}

	for _, inv := range inventory {
		_, err := db.Exec(`
			INSERT OR IGNORE INTO inventory (book_id, in_stock, quantity, warehouse, shipping_time) 
			VALUES (?, ?, ?, ?, ?)
		`, inv["book_id"], inv["in_stock"], inv["quantity"], inv["warehouse"], inv["shipping_time"])
		if err != nil {
			return err
		}
	}

	// Insert reviews data
	reviews := []map[string]interface{}{
		{"book_id": "1", "average_rating": 4.5, "total_reviews": 89, "recent_review": "Essential reading for Go developers", "five_star": 45, "four_star": 28, "three_star": 12, "two_star": 3, "one_star": 1},
		{"book_id": "2", "average_rating": 4.3, "total_reviews": 127, "recent_review": "Changed how I think about writing code", "five_star": 65, "four_star": 32, "three_star": 20, "two_star": 7, "one_star": 3},
		{"book_id": "3", "average_rating": 4.7, "total_reviews": 56, "recent_review": "Incredibly helpful for interview prep", "five_star": 38, "four_star": 14, "three_star": 3, "two_star": 1, "one_star": 0},
		{"book_id": "4", "average_rating": 4.1, "total_reviews": 94, "recent_review": "Eye-opening perspective on modern life", "five_star": 42, "four_star": 31, "three_star": 15, "two_star": 4, "one_star": 2},
	}

	for _, rev := range reviews {
		_, err := db.Exec(`
			INSERT OR IGNORE INTO reviews (book_id, average_rating, total_reviews, recent_review, five_star, four_star, three_star, two_star, one_star) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, rev["book_id"], rev["average_rating"], rev["total_reviews"], rev["recent_review"], rev["five_star"], rev["four_star"], rev["three_star"], rev["two_star"], rev["one_star"])
		if err != nil {
			return err
		}
	}

	return nil
}

// Database query functions for fetching book information

// FetchBookMetadata retrieves basic book information from the books table
func FetchBookMetadata(bookID string) map[string]interface{} {
	var title, author, isbn, publishDate, description string

	err := db.QueryRow(`
		SELECT title, author, isbn, publish_date, description 
		FROM books 
		WHERE id = ?
	`, bookID).Scan(&title, &author, &isbn, &publishDate, &description)

	if err != nil {
		log.Printf("Error fetching book metadata for ID %s: %v", bookID, err)
		return map[string]interface{}{
			"error": "Failed to fetch book metadata",
		}
	}

	return map[string]interface{}{
		"title":        title,
		"author":       author,
		"isbn":         isbn,
		"publish_date": publishDate,
		"description":  description,
	}
}

// FetchBookPricing retrieves pricing information from the pricing table
func FetchBookPricing(bookID string) map[string]interface{} {
	var price, discount, salePrice float64
	var currency, promotion string

	err := db.QueryRow(`
		SELECT price, currency, discount, sale_price, promotion 
		FROM pricing 
		WHERE book_id = ?
	`, bookID).Scan(&price, &currency, &discount, &salePrice, &promotion)

	if err != nil {
		log.Printf("Error fetching book pricing for ID %s: %v", bookID, err)
		return map[string]interface{}{
			"error": "Failed to fetch pricing information",
		}
	}

	return map[string]interface{}{
		"price":      price,
		"currency":   currency,
		"discount":   discount,
		"sale_price": salePrice,
		"promotion":  promotion,
	}
}

// FetchBookInventory retrieves inventory status from the inventory table
func FetchBookInventory(bookID string) map[string]interface{} {
	var inStock bool
	var quantity int
	var warehouse, shippingTime string

	err := db.QueryRow(`
		SELECT in_stock, quantity, warehouse, shipping_time 
		FROM inventory 
		WHERE book_id = ?
	`, bookID).Scan(&inStock, &quantity, &warehouse, &shippingTime)

	if err != nil {
		log.Printf("Error fetching book inventory for ID %s: %v", bookID, err)
		return map[string]interface{}{
			"error": "Failed to fetch inventory information",
		}
	}

	return map[string]interface{}{
		"in_stock":      inStock,
		"quantity":      quantity,
		"warehouse":     warehouse,
		"shipping_time": shippingTime,
	}
}

// FetchBookReviews retrieves customer review data from the reviews table
func FetchBookReviews(bookID string) map[string]interface{} {
	var averageRating float64
	var totalReviews, fiveStar, fourStar, threeStar, twoStar, oneStar int
	var recentReview string

	err := db.QueryRow(`
		SELECT average_rating, total_reviews, recent_review, five_star, four_star, three_star, two_star, one_star 
		FROM reviews 
		WHERE book_id = ?
	`, bookID).Scan(&averageRating, &totalReviews, &recentReview, &fiveStar, &fourStar, &threeStar, &twoStar, &oneStar)

	if err != nil {
		log.Printf("Error fetching book reviews for ID %s: %v", bookID, err)
		return map[string]interface{}{
			"error": "Failed to fetch reviews",
		}
	}

	return map[string]interface{}{
		"average_rating": averageRating,
		"total_reviews":  totalReviews,
		"recent_review":  recentReview,
		"rating_breakdown": map[string]int{
			"5_star": fiveStar,
			"4_star": fourStar,
			"3_star": threeStar,
			"2_star": twoStar,
			"1_star": oneStar,
		},
	}
}

// FetchPersonalizedRecommendations - Simple external API call example
func FetchPersonalizedRecommendations(bookID string, userID string) map[string]interface{} {
	// Step 1: Make a simple external API call to get a random quote
	response, err := httpClient.Get("https://zenquotes.io/api/random")

	// Step 2: Handle network errors
	if err != nil {
		log.Printf("Error calling external API: %v", err)
		return map[string]interface{}{
			"error":  "Failed to fetch recommendations",
			"source": "external_api_failed",
		}
	}
	defer response.Body.Close() // Always close the response body!

	// Step 3: Parse the JSON response
	var quoteData []map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&quoteData)
	if err != nil {
		log.Printf("Error parsing API response: %v", err)
		return map[string]interface{}{
			"error": "Failed to parse API response",
		}
	}

	// Step 4: Use the external data in your response
	return map[string]interface{}{
		"user_id":        userID,
		"book_id":        bookID,
		"external_quote": quoteData, // This is real data from the external API!
		"recommendations": []map[string]interface{}{
			{
				"title":  "Based on your reading preferences...",
				"source": "external_api_enriched",
			},
		},
		"api_source": "zenquotes.io",
	}
}
