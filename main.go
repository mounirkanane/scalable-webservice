package main

import "net/http"
import "log"
import "encoding/json"

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
	{ID: "4", Title: "Dopamine Nation", Author: "Anne Lembke", Price: 20.00},
}

func main() {
	http.HandleFunc("/api/books", booksHandler)
	log.Println("Starting server on http://localhost:8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("FATAL: error while starting server:", err)
    }
}

func booksHandler(w http.ResponseWriter, r *http.Request) {
	// Validate the HTTP method
	if r.Method != http.MethodGet {
		log.Printf("Method %s not allowed for %s", r.Method, r.URL.Path)
		// Sets status code and sends an error message
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) // Client errors are 400s, in this case 405
		return                                                           // Need this to prevent rest of handler from executing after error
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Encode and stream books as a JSON response, the JSON is sent across the network as it's generated
	err := json.NewEncoder(w).Encode(books)
	if err != nil {
		log.Printf("Error occured while encoding JSON: %v", err)
	}

	// Log successful operation
	log.Printf("Successfully returned %d books to %s", len(books), r.RemoteAddr)
}
