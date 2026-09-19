package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type App struct {
	db   *sql.DB
	tmpl *template.Template
}

// ParseTemplates loads every page template with the shared function map
// (formatDate) available to all of them.
func ParseTemplates() (*template.Template, error) {
	return template.New("").Funcs(template.FuncMap{
		"formatDate": FormatDateDisplay,
	}).ParseFiles(
		"templates/index.html",
		"templates/vehicle.html",
		"templates/new_vehicle.html",
		"templates/edit_vehicle.html",
		"templates/new_record.html",
		"templates/record.html",
		"templates/error.html",
	)
}

// Remembar that here we can eventually create a function like
// func (a *App) NewRouter() http.Handler {}

func main() {
	r := chi.NewRouter()

	template, parse_err := ParseTemplates()
	if parse_err != nil {
		log.Fatal(parse_err)
	}

	db, err := InitDBandCreateOrOpenTables()
	if err != nil {
		log.Fatal("Error with database: ", err)
	}
	defer db.Close()
	app := App{db, template}

	// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("Welcome!"))
	// })
	r.Get("/", app.IndexHandler)
	r.Get("/vehiculos/{id}", app.VehicleHandler)
	r.Post("/vehiculos/{id}", app.VehicleHandler)
	r.Get("/vehiculos/new", app.NewVehicleHandler)
	r.Post("/vehiculos/new", app.NewVehicleHandler)

	log.Println("Server running and listening on localhost:8081")
	_ = http.ListenAndServe(":8081", r)
}
