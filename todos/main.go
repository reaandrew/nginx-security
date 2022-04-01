package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Todo struct{
	ID int `json:"id"`
	Value string `json:"value"`
}

type TodoCollection struct{
	mu       sync.Mutex
	count int
	todos []Todo
}

func (collection *TodoCollection) Create(value string) Todo{
	collection.mu.Lock()
	defer collection.mu.Unlock()
	collection.count++
	todo :=  Todo{
		ID: collection.count,
		Value: value,
	}
	collection.todos = append(collection.todos,todo)
	return todo
}

func (collection *TodoCollection) Update(id int, value string){
	collection.mu.Lock()
	defer collection.mu.Unlock()
	for index, todo := range collection.todos{
		if todo.ID == id{
			collection.todos[index].Value = value
		}
		break
	}
}

func (collection *TodoCollection) List() []Todo{
	return collection.todos
}

// logging is middleware for wrapping any handler we want to track response
// times for and to see what resources are requested.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		req := fmt.Sprintf("%s %s", r.Method, r.URL)
		log.Println(req)
		next.ServeHTTP(w, r)
		log.Println(req, "completed in", time.Now().Sub(start))
	})
}

// index is the handler responsible for rending the index page for the site.
func todos(collection TodoCollection) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET"{
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(collection.List())
		}
		if r.Method == "POST"{
			decoder := json.NewDecoder(r.Body)
			var data Todo
			err := decoder.Decode(&data)
			if err != nil {
				panic(err)
			}
			todo := collection.Create(data.Value)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("ETag", strconv.Itoa(todo.ID))
			w.WriteHeader(http.StatusCreated)
		}
		if r.Method == "PUT"{
			decoder := json.NewDecoder(r.Body)
			var data Todo
			err := decoder.Decode(&data)
			if err != nil {
				panic(err)
			}
			collection.Update(data.ID, data.Value)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNoContent)
		}
	})
}

func main(){
	collection := TodoCollection{todos: []Todo{}}
	mux := http.NewServeMux()
	mux.Handle("/todos", logging(todos(collection)))

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	server := http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	log.Println("main: running todo server on port", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("main: couldn't start todo server: %v\n", err)
	}
}