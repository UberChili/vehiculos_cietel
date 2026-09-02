package main

import (
	"database/sql"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func CreateOrOpenTable() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./vehicles.db?_foreign_keys=on")
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
			last_service TEXT DEFAULT '',
			photo_url TEXT DEFAULT ''
			);
		`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	} else {
		log.Println("Table 'vehicles' created successfully")
	}

	// Migration for dbs created before photo_url existed.
	if _, err := db.Exec("ALTER TABLE vehicles ADD COLUMN photo_url TEXT DEFAULT ''"); err != nil &&
		!strings.Contains(err.Error(), "duplicate column name") {
		return nil, err
	}

	recordsStmt := `
				CREATE TABLE IF NOT EXISTS records (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				vehicle_id INTEGER NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
				date TEXT NOT NULL,
				type TEXT NOT NULL,
				description TEXT NOT NULL,
				cost TEXT DEFAULT ''
				);
			`

	if _, err = db.Exec(recordsStmt); err != nil {
		return nil, err
	} else {
		log.Println("Table 'records' created successfully")
	}

	return db, nil
}

func (a *App) InsertVehicle(vehicle Vehicle) error {
	if err := vehicle.Validate(); err != nil {
		return err
	}
	// Check for valid plate (at least length of chars)
	if len(vehicle.Plate) > 10 {
		return &ValidationError{"La placa es demasiado larga (máximo 10 caracteres)"}
	}
	_, err := a.db.Exec("INSERT INTO vehicles(plate, maker, model, year, assigned_to, location, last_service, photo_url) VALUES(?, ?, ?, ?, ?, ?, ?, ?)",
		vehicle.Plate, vehicle.Maker, vehicle.Model, vehicle.Year, vehicle.AssignedTo, vehicle.Location, vehicle.LastService, vehicle.PhotoURL)
	if err != nil {
		return err
	}
	log.Println("New vehicle inserted successfully")

	return nil
}

func (a *App) InsertRecord(vehicle_id string, record Record) error {
	if err := record.Validate(); err != nil {
		return err
	}

	_, err := a.db.Exec("INSERT INTO records(vehicle_id, date, type, description, cost) VALUES(?, ?, ?, ?, ?)",
		vehicle_id, record.DateShort, record.Type, record.Description, record.Cost)
	if err != nil {
		return err
	}
	log.Printf("New record for vehicle %s inserted succesfully\n", vehicle_id)
	return nil
}
