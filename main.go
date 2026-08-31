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

// var templ *template.Template

func main() {
	var err error
	templ, err := template.ParseFiles("templates/index.html", "templates/vehicle.html", "templates/new_vehicle.html", "templates/error.html")
	if err != nil {
		log.Fatal(err)
	}

	db, err := CreateOrOpenTable()
	if err != nil {
		log.Fatal("Could not Open or Create table: ", err)
	}
	app := App{db, templ}
	defer app.db.Close()

	// Using Standard Library
	// http.HandleFunc("/vehiculos/", app.HomeHandler)
	// http.ListenAndServe(":8080", nil)

	// Using Chi
	r := chi.NewRouter()
	r.Get("/vehiculos/", app.HomeHandler)
	r.Get("/vehiculos/{id}", app.HomeHandler)
	r.Get("/vehiculos/new/", app.NewVehicleHandler)
	r.Get("/vehiculos/new", app.NewVehicleHandler)
	r.Post("/vehiculos/new/", app.NewVehicleHandler)
	r.Post("/vehiculos/new", app.NewVehicleHandler)

	http.ListenAndServe(":8080", r)
}
