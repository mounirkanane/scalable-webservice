package main

// Book represents the basic book structure for the books list endpoint
type Book struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
}

// BookDetailsResponse represents the comprehensive book details response
type BookDetailsResponse struct {
	BookID          string                 `json:"book_id"`
	Metadata        map[string]interface{} `json:"metadata"`
	Pricing         map[string]interface{} `json:"pricing"`
	Inventory       map[string]interface{} `json:"inventory"`
	Reviews         map[string]interface{} `json:"reviews"`
	Recommendations map[string]interface{} `json:"recommendations"`
	Duration        int64                  `json:"duration"`
}

// In-memory books data for the simple books list endpoint
var books = []Book{
	{ID: "1", Title: "The Go Programming Language", Author: "Alan Donovan", Price: 39.99},
	{ID: "2", Title: "Clean Code", Author: "Robert Martin", Price: 32.50},
	{ID: "3", Title: "System Design Interview", Author: "Alex Xu", Price: 28.95},
	{ID: "4", Title: "Dopamine Nation", Author: "Anna Lembke", Price: 20.00},
}
