package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Vehicle struct {
	ID          int
	Maker       string
	Model       string
	Year        string
	Plate       string
	LastService string
	LastRepair  string
	AssignedTo  string
	Location    string
	Records     []string
	PhotoURL    string
}

var vehicles = []Vehicle{
	{1, "Chevrolet", "Spark", "2018", "ABC-123", "21-08-2026", "", "Abraham", "Morelia", nil, ""},
	{2, "Chevrolet", "Chevy", "2010", "DEF-456", "05-06-2026", "", "Miguel", "Pátzcuaro", nil, ""},
	{3, "Hyundai", "Atos", "2011", "GHI-789", "31-07-2026", "", "Fulano", "Vallarta", nil, ""},
	{4, "Ford", "Icon", "2015", "JKL-123", "30-05-2026", "", "Hermilo", "Morelia", nil, ""},
}

func HomeHandler(w http.ResponseWriter, req *http.Request) {
	// for _, v := range vehicles {
	// 	fmt.Println("Vehicle: %s\t%s\t%s\t%s\n", v.Maker, v.Model, v.Year, v.Plate)
	// }

	id := strings.TrimPrefix(req.URL.Path, "/vehiculos/")
	if id == "" {
		err := templ.ExecuteTemplate(w, "index.html", vehicles)
		if err != nil {
			log.Fatal(err)
		}
		return
	} else {
		id, err := strconv.Atoi(id)
		if err != nil {
			fmt.Fprintf(w, "Error when converting id to int: %s\n", err)
			return
		}
		for i, vehicle := range vehicles {
			if vehicle.ID == id {
				err = templ.ExecuteTemplate(w, "vehicle.html", vehicles[i])
				if err != nil {
					fmt.Printf("Could not execute tempalte %s\n", err)
					return
				}
			}
		}
	}
}

func VehicleHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		id := strings.TrimPrefix(req.URL.Path, "/vehiculos/")
		if id == "" {
			return
		} else {
			id, err := strconv.Atoi(id)
			if err != nil {
				fmt.Fprintf(w, "Error when converting id to int: %s\n", err)
				return
			}
			fmt.Println("ID: ", id)
			err = templ.Execute(w, vehicles)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		fmt.Println("POST request")
	}
}

var templ *template.Template

func main() {
	var err error
	templ, err = template.ParseFiles("templates/index.html", "templates/vehicle.html")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/vehiculos/", HomeHandler)
	http.ListenAndServe(":8080", nil)
}
