package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/handlers"
)

const PORT = 8080

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Serve the public/ folder at all times
	workdir, _ := os.Getwd()
	fmt.Println("Working dir:", workdir)
	filesDir := http.Dir(filepath.Join(workdir, "public"))
	handlers.FileServer(r, "/public", filesDir)

	// Routes
	r.Get("/", handlers.IndexHandler{}.Show)
	r.Get("/home", handlers.HomeHandler{}.Show)

	fmt.Printf("Listening on port :%d\n", PORT)
	http.ListenAndServe(fmt.Sprintf(":%d", PORT), r)
}
