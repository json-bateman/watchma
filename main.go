package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

const PORT = 8080

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	fmt.Printf("Listening on port :%d\n", PORT)
	http.ListenAndServe(fmt.Sprintf(":%d", PORT), r)
}
