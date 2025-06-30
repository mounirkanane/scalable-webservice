package main

import "net/http"

type book struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
}

var books = []book{
	{ID: "1", Title: "The Go Programming Language", Author: "Alan Donovan", Price: 39.99},
	{ID: "2", Title: "Clean Code", Author: "Robert Martin", Price: 32.50},
	{ID: "3", Title: "System Design Interview", Author: "Alex Xu", Price: 28.95},
}

func main() {
	http.HandleFunc("/api/books", booksHandler)
}

func booksHandler(w http.ResponseWriter, r *http.Request) {
	// Validate the HTTP method
	if r.Method != http.MethodGet {
		// Sets status code and sends an error message
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) // Client errors are 400s, in this case 405
		return                                                           // Need this to prevent rest of handler from executing after error
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")
}
