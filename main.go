package main

import (
	"html/template"
	"log"
	"net/http"
)

type Vehicle struct {
	ID          int
	Maker       string
	Model       string
	Year        string
	Plate       string
	LastService string
	AssignedTo  string
	Location    string
}

var vehicles = []Vehicle{
	{1, "Chevrolet", "Spark", "2018", "ABC-123", "21-08-2026", "Abraham", "Morelia"},
	{2, "Chevrolet", "Chevy", "2010", "DEF-456", "05-06-2026", "Miguel", "Pátzcuaro"},
	{3, "Hyndai", "Atos", "2011", "GHI-789", "31-07-2026", "Fulano", "Vallarta"},
	{4, "Ford", "Icon", "2015", "JKL-123", "30-05-2026", "Hermilo", "Morelia"},
}

func HomeHandler(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	// for _, v := range vehicles {
	// 	fmt.Println("Vehicle: %s\t%s\t%s\t%s\n", v.Maker, v.Model, v.Year, v.Plate)
	// }
	err = tmpl.Execute(w, vehicles)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/vehiculos", HomeHandler)
	http.ListenAndServe(":8080", nil)
}
