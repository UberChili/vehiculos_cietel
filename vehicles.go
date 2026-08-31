package main

import (
	"database/sql"
	"net/http"
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

func NewVehicleFromForm(r *http.Request) Vehicle {
	// process and sanitize strings
	plate := strings.ToUpper(r.PostFormValue("plate"))
	maker := CapitalizeFirst(strings.ToLower(r.PostFormValue("maker")))
	model := CapitalizeFirst(strings.ToLower(r.PostFormValue("model")))
	year := r.PostFormValue("year")
	assigned_to := r.PostFormValue("assigned_to")
	if assigned_to != "" {
		assigned_to = CapitalizeFirst(strings.ToLower(assigned_to))
	}
	location := r.PostFormValue("location")
	if location != "" {
		location = CapitalizeFirst(strings.ToLower(location))
	}
	new_vehicle := Vehicle{Maker: maker, Model: model, Year: year, Plate: plate, AssignedTo: assigned_to, Location: location}

	return new_vehicle
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
