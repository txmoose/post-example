package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var err error

type DbEntry struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func InitializeDB() {
	log.Print("Initializing Database")
	DB, err = gorm.Open(sqlite.Open("prod.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	log.Print("Migrating DB Schema")
	// Migrate the schema
	err = DB.AutoMigrate(&DbEntry{})
	if err != nil {
		log.Fatal("Failed to migrate database")
	}

	DB.Create(&DbEntry{Key: "initial-test", Value: "initial-value"})
}

func main() {
	InitializeDB()

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	log.Print("Initializing Routes")
	r := mux.NewRouter()
	// Add your routes as needed
	r.HandleFunc("/create", CreateHandler).Methods("POST")
	r.HandleFunc("/get/{key}", GetHandler).Methods("GET")

	log.Print("Starting server")
	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request to CreateHandler - [%v]", r.RemoteAddr)
	switch r.Method {
	case http.MethodPost:
		var NewEntry DbEntry
		err := json.NewDecoder(r.Body).Decode(&NewEntry)
		if err != nil {
			log.Printf("Error decoding sent JSON - %v", r.Body)
			http.Error(w, "Error decoding Sent JSON", http.StatusBadRequest)
			return
		}
		log.Printf("Key - [%s] === Value - [%s]", NewEntry.Key, NewEntry.Value)

		log.Print("Creating Database Entry")
		DB.Create(&NewEntry)
		log.Print("Successfully created Database Entry")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Success!"))
		return

	default:
		log.Print("Request had incorrect method")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request to GetHandler - [%v]", r.Host)
	switch r.Method {
	case http.MethodGet:
		var GetEntry DbEntry
		params := mux.Vars(r)
		log.Printf("Params are - %s", params["key"])
		log.Printf("DB -> [%s]", DB.Name())
		DB.Where("key = ? ", params["key"]).First(&GetEntry)
		log.Printf("GetEntry has key - [%s] value - [%s]", GetEntry.Key, GetEntry.Value)
		err := json.NewEncoder(w).Encode(GetEntry)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")

	default:
		log.Print("Request had incorrect method")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
