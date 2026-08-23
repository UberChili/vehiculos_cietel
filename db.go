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
			assigned_to TEXT,
			location TEXT,
			last_service TEXT
			);
		`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	log.Println("Table 'vehicles' created successfully")
	return db, nil
}

// func OpenTable() (*sql.DB, error) {
// 	db, err := sql.Open("sqlite3", "./vehicles.db")
// 	if err != nil {
// 		return nil, err
// 	}

// 	return db, nil
// }

func AddVehicle(db *sql.DB, vehicle Vehicle) error {
	_, err := db.Exec("INSERT INTO vehicles(plate, maker, model, year, assigned_to, location, last_service) VALUES(?, ?, ?, ?, ?, ?, ?)",
		vehicle.Plate, vehicle.Maker, vehicle.Model, vehicle.Year, vehicle.AssignedTo, vehicle.Location, vehicle.AssignedTo)
	if err != nil {
		return err
	}
	log.Println("New vehicle inserted successfully")

	return nil
}
