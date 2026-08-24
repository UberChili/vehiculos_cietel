package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

type App struct {
	db   *sql.DB
	tmpl *template.Template
}

// var templ *template.Template

func main() {
	var err error
	templ, err = template.ParseFiles("templates/index.html", "templates/vehicle.html")
	if err != nil {
		log.Fatal(err)
	}

	db, err := CreateOrOpenTable()
	if err != nil {
		log.Fatal("Could not Open or Create table: ", err)
	}
	app := App{db, templ}
	defer app.db.Close()

	http.HandleFunc("/vehiculos/", app.HomeHandler)
	http.ListenAndServe(":8080", nil)
}
