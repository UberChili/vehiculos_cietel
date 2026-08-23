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

	db, err := CreateOrOpenTable()
	if err != nil {
		log.Fatal("Could not Open or Create table: ", err)
	}
	defer db.Close()

	http.HandleFunc("/vehiculos/", HomeHandler)
	http.ListenAndServe(":8080", nil)
}
