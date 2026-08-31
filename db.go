package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func CreateOrOpenTable() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./vehicles.db")
	if err != nil {
		return nil, err
	}

	sqlStmt := `
			CREATE TABLE IF NOT EXISTS vehicles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			plate TEXT NOT NULL,
			maker TEXT NOT NULL,
			model TEXT NOT NULL,
			year TEXT NOT NULL,
			assigned_to TEXT DEFAULT '',
			location TEXT DEFAULT '',
			last_service TEXT DEFAULT ''
			);
		`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	} else {
		log.Println("Table 'vehicles' created successfully")
	}
	return db, nil
}

func (a *App) InsertVehicle(vehicle Vehicle) error {
	_, err := a.db.Exec("INSERT INTO vehicles(plate, maker, model, year, assigned_to, location, last_service) VALUES(?, ?, ?, ?, ?, ?, ?)",
		vehicle.Plate, vehicle.Maker, vehicle.Model, vehicle.Year, vehicle.AssignedTo, vehicle.Location, vehicle.LastService)
	if err != nil {
		return err
	}
	log.Println("New vehicle inserted successfully")

	return nil
}
