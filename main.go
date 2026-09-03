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

// NewRouter wires up every route on top of app. Split out from main so
// tests can exercise the exact same routing table via httptest, instead of
// duplicating (and risking drift from) the route list.
func NewRouter(app *App) http.Handler {
	r := chi.NewRouter()
	r.Get("/vehiculos", app.HomeHandler)
	r.Get("/vehiculos/", app.HomeHandler)
	r.Get("/vehiculos/{id}", app.HomeHandler)
	r.Get("/vehiculos/{id}/", app.HomeHandler)
	r.Get("/vehiculos/{id}/nuevo-registro", app.NewRecordHandler)
	r.Get("/vehiculos/{id}/nuevo-registro/", app.NewRecordHandler)
	r.Post("/vehiculos/{id}/nuevo-registro/", app.NewRecordHandler)
	r.Post("/vehiculos/{id}/nuevo-registro", app.NewRecordHandler)
	r.Get("/vehiculos/{id}/registro/{record_id}", app.RecordHandler)
	r.Get("/vehiculos/{id}/registro/{record_id}/", app.RecordHandler)
	r.Post("/vehiculos/{id}/registro/{record_id}/eliminar", app.DeleteRecordHandler)
	r.Post("/vehiculos/{id}/registro/{record_id}/eliminar/", app.DeleteRecordHandler)
	r.Get("/vehiculos/{id}/editar", app.EditVehicleHandler)
	r.Get("/vehiculos/{id}/editar/", app.EditVehicleHandler)
	r.Post("/vehiculos/{id}/editar", app.EditVehicleHandler)
	r.Post("/vehiculos/{id}/editar/", app.EditVehicleHandler)
	r.Get("/vehiculos/new/", app.NewVehicleHandler)
	r.Get("/vehiculos/new", app.NewVehicleHandler)
	r.Post("/vehiculos/new/", app.NewVehicleHandler)
	r.Post("/vehiculos/new", app.NewVehicleHandler)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	return r
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

func main() {
	templ, err := ParseTemplates()
	if err != nil {
		log.Fatal(err)
	}

	db, err := CreateOrOpenTable("./vehicles.db")
	if err != nil {
		log.Fatal("Could not Open or Create table: ", err)
	}
	app := App{db, templ}
	defer app.db.Close()

	r := NewRouter(&app)
	log.Fatal(http.ListenAndServe("127.0.0.1:8081", r))
}
