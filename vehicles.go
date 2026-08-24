package main

import (
	"fmt"
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

func (a *App) HomeHandler(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/vehiculos/")

	if id == "" {
		// Display all vehicles on main page
		// rows := "something"
		err := a.tmpl.ExecuteTemplate(w, "index.html", vehicles)
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
				err = a.tmpl.ExecuteTemplate(w, "vehicle.html", vehicles[i])
				if err != nil {
					fmt.Printf("Could not execute tempalte %s\n", err)
					return
				}
			}
		}
	}
}

func NewVehicleHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Hello there. This is the handler that would add a new vehicle!")
}
