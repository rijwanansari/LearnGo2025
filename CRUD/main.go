package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

type Customer struct{}

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
	DeletedAt   time.Time `json:"deleted_at"`
}

var db *pgx.Conn

func main() {

	// connection string for PostgreSQL
	dsn := "postgres://username:password@localhost:5432/CsvWorkSync"

	var err error
	db, err = pgx.Connect(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Connected to database successfully!")

	defer db.Close(context.Background())

	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/event", eventHandler)
	http.HandleFunc("/events", eventsHandler)

	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", nil)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

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

func eventHandler(w http.ResponseWriter, r *http.Request) {

}

func listEventHandler(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query(context.Background(), "SELECT id, name, description, location, start_time, end_time, organizer, attendees, created_at, updated_at FROM events")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.ID, &event.Name, &event.Description, &event.Location, &event.StartTime, &event.EndTime, &event.Organizer, &event.Attendees, &event.CreatedAt, &event.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		events = append(events, event)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func createEventHandler(w http.ResponseWriter, r *http.Request) {
	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(event)

	var id int
	err := db.QueryRow(context.Background(), "INSERT INTO events (name, description, location, start_time, end_time, organizer, attendees) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		event.Name, event.Description, event.Location, event.StartTime, event.EndTime, event.Organizer, event.Attendees).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Event created with ID: %d", id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]interface{}{
		"id":          id,
		"name":        event.Name,
		"description": event.Description,
		"location":    event.Location,
		"start_time":  event.StartTime,
		"end_time":    event.EndTime,
		"organizer":   event.Organizer,
		"attendees":   event.Attendees,
		"created_at":  event.CreatedAt,
		"updated_at":  event.UpdatedAt,
		"deleted_at":  event.DeletedAt,
	}
	json.NewEncoder(w).Encode(response)
}

func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(&struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Alias
	}{
		StartTime: e.StartTime.Format(time.RFC3339),
		EndTime:   e.EndTime.Format(time.RFC3339),
		Alias:     (Alias)(e),
	})
}
