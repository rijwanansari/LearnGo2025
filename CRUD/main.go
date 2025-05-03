package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

type Event struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Organizer   string    `json:"organizer"`
	Attendees   string    `json:"attendees"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at,omitempty"`
}

var db *pgx.Conn

// Main function - Entry point for the program
func main() {
	// Initialize DB connection
	initDB()

	// Setup HTTP routes
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/events", eventsHandler)

	// Start HTTP server
	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// initDB initializes the PostgreSQL connection
func initDB() {
	dsn := getEnv("DATABASE_URL", "postgres://username:password@localhost:5432/CsvWorkSync")

	var err error
	db, err = pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	log.Println("Connected to database successfully!")
}

// getEnv is a helper function to fetch environment variables with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// helloHandler handles requests to the root URL
func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello, World!")
}

// eventsHandler routes events-related actions (GET, POST)
func eventsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listEventHandler(w, r)
	case http.MethodPost:
		createEventHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listEventHandler handles fetching a list of events
func listEventHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(context.Background(), "SELECT id, name, description, location, start_time, end_time, organizer, attendees, created_at, updated_at, deleted_at FROM events")
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.ID, &event.Name, &event.Description, &event.Location, &event.StartTime, &event.EndTime, &event.Organizer, &event.Attendees, &event.CreatedAt, &event.UpdatedAt, &event.DeletedAt); err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}
		events = append(events, event)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(events); err != nil {
		handleError(w, err, http.StatusInternalServerError)
	}
}

// createEventHandler handles event creation via POST request
func createEventHandler(w http.ResponseWriter, r *http.Request) {
	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var id int
	err := db.QueryRow(context.Background(), "INSERT INTO events (name, description, location, start_time, end_time, organizer, attendees) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		event.Name, event.Description, event.Location, event.StartTime, event.EndTime, event.Organizer, event.Attendees).Scan(&id)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	event.ID = id // Update event struct with the newly created ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(event); err != nil {
		handleError(w, err, http.StatusInternalServerError)
	}
}

// handleError is a helper function to handle errors in a standardized way
func handleError(w http.ResponseWriter, err error, statusCode int) {
	http.Error(w, fmt.Sprintf("Error: %v", err), statusCode)
	log.Printf("Error: %v", err)
}

// MarshalJSON customizes the JSON encoding for the Event struct
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(&struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		DeletedAt string `json:"deleted_at,omitempty"`
		Alias
	}{
		StartTime: e.StartTime.Format(time.RFC3339),
		EndTime:   e.EndTime.Format(time.RFC3339),
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
		DeletedAt: formatNullableTime(e.DeletedAt),
		Alias:     (Alias)(e),
	})
}

// formatNullableTime formats nullable timestamps
func formatNullableTime(t time.Time) string {
	if t.IsZero() {
		return "" // Return empty string for "zero" time values (like NULL)
	}
	return t.Format(time.RFC3339)
}
