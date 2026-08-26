package main

import (
	"database/sql"
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

func (a *App) NewVehicleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		err := a.tmpl.ExecuteTemplate(w, "new_vehicle.html", nil)
		if err != nil {
			return
		}
	}

	if r.Method == "POST" {
		// Add vehicle to db
		// TODO
		// Now redirect
		http.Redirect(w, r, "/vehiculos/", http.StatusSeeOther)
		return
	}
}

func (a *App) HomeHandler(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/vehiculos/")

	// If there's no prefix, only 'vehiculos' was called, so we only need the index list
	if id == "" {
		// Get all vehicles from database
		vehicles, err := a.GetVehicles()
		if err != nil {
			// We don't have cars?
			// Was one car malformed or incomplete?
			log.Fatal("Error when querying for all cars: ", err)
		}
		// Execute template with cars
		err = a.tmpl.ExecuteTemplate(w, "index.html", vehicles)
		if err != nil {
			log.Fatal(err)
		}
		return
	} else {
		// User clicked on a car, so we need to obtain a specific car information
		// And call the vehicle template with that specific car info
		id, err := strconv.Atoi(id)
		if err != nil {
			fmt.Fprintf(w, "Error when converting id to int: %s\n", err)
			return
		}

		vehicle, err := a.GetVehicleByID(id)
		if err != nil {
			log.Fatal("Could not get vehicle from database: ", err)
		}
		err = a.tmpl.ExecuteTemplate(w, "vehicle.html", vehicle)
		if err != nil {
			fmt.Printf("Could not execute tempalte %s\n", err)
			return
		}
	}
}

func NewVehicleHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Hello there. This is the handler that would add a new vehicle!")
}

func (a *App) GetVehicleByID(id int) (Vehicle, error) {
	var v Vehicle
	row := a.db.QueryRow("SELECT id, plate, maker, model, year, assigned_to, location, last_service FROM vehicles WHERE id = ?", id)
	if row.Err() == sql.ErrNoRows {
		return v, row.Err()
	}
	err := row.Scan(&v.ID, &v.Plate, &v.Maker, &v.Model, &v.Year, &v.AssignedTo,
		&v.Location, &v.LastService)
	if err != nil {
		return v, err
	}
	return v, nil
}

func (a *App) GetVehicles() ([]Vehicle, error) {
	rows, err := a.db.Query("SELECT id, plate, maker, model, year, assigned_to, location, last_service FROM vehicles")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []Vehicle

	// Loop through rows, using Scan to assign column data to struct fields
	for rows.Next() {
		var v Vehicle
		if err := rows.Scan(&v.ID, &v.Plate, &v.Maker, &v.Model, &v.Year, &v.AssignedTo,
			&v.Location, &v.LastService); err != nil {
			return vehicles, err
		}
		vehicles = append(vehicles, v)
	}
	if err = rows.Err(); err != nil {
		return vehicles, err
	}

	return vehicles, nil
}
