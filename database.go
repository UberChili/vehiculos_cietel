package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

func InitDBandCreateOrOpenTables() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "vehicles.db?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	log.Println("Successfully connected to SQLite database.")

	// Ensuring required databases exist
	log.Println("Opening or creating tables...")
	technicians_table_stmt := `CREATE TABLE IF NOT EXISTS technicians (id INTEGER PRIMARY KEY AUTOINCREMENT, first_name TEXT NOT NULL, last_name TEXT NOT NULL);`
	vehicles_table_stmt := `CREATE TABLE IF NOT EXISTS vehicles (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				plate TEXT NOT NULL,
				maker TEXT NOT NULL,
				model TEXT NOT NULL,
				year TEXT NOT NULL,
				assigned_to INTEGER UNIQUE REFERENCES technicians(id),
				location TEXT DEFAULT '',
				last_service TEXT DEFAULT '',
				photo_url TEXT DEFAULT '');`
	records_table_stmt := `CREATE TABLE IF NOT EXISTS records (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				vehicle_id INTEGER NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
				date TEXT NOT NULL,
				type TEXT NOT NULL,
				description TEXT NOT NULL,
				cost TEXT DEFAULT ''
				);`
	_, err = db.Exec(technicians_table_stmt)
	if err != nil {
		return nil, fmt.Errorf("error creating technicians table: %w", err)
	}
	_, err = db.Exec(vehicles_table_stmt)
	if err != nil {
		return nil, fmt.Errorf("error creating vehicles table: %w", err)
	}
	_, err = db.Exec(records_table_stmt)
	if err != nil {
		return nil, fmt.Errorf("error creating records table: %w", err)
	}

	log.Println("Succesfully opened tables.")
	return db, nil
}

func (a App) findAllVehicles() ([]Vehicle, error) {
	query := `SELECT id, plate, maker, model, year, assigned_to, location FROM vehicles ORDER BY LOCATION`

	rows, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []Vehicle

	for rows.Next() {
		v := &Vehicle{}
		err := rows.Scan(
			&v.ID, &v.Plate, &v.Maker, &v.Model, &v.Year, &v.AssignedTo, &v.Location)

		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, *v)
	}
	return vehicles, nil
}

func (a App) GetVehicle(id string) (Vehicle, error) {
	row := a.db.QueryRow("SELECT id, plate, maker, model, year, assigned_to, location FROM vehicles WHERE id = ?", id)

	v := Vehicle{}

	err := row.Scan(&v.ID, &v.Plate, &v.Maker, &v.Model, &v.Year, &v.AssignedTo,
		&v.Location)

	// Not neccessarily an error, but no results
	if err == sql.ErrNoRows {
		return Vehicle{}, errors.New("Vehicle not found")
	}

	records_rows, err := a.db.Query("SELECT id, date, type, description, cost FROM records WHERE vehicle_id = ?", id)
	if err != nil {
		v.Records = nil
		return v, err
	}

	for records_rows.Next() {
		r := &Record{}
		err := records_rows.Scan(&r.ID, &r.DateShort, &r.Type, &r.Description, &r.Cost)

		if err != nil {
			continue
		}
		v.Records = append(v.Records, *r)
	}

	return v, nil
}

func (a App) GetLastServiceOrRepair(vehicle Vehicle) Record {
	row := a.db.QueryRow("SELECT * FROM records WHERE id = ? ORDER BY DATE", vehicle.ID)

	var r Record
	err := row.Scan(&r.ID, &r.DateShort, &r.Type, &r.Description, &r.Cost)
	if err == sql.ErrNoRows {
		log.Printf("No records found for vehicle_id %d: %s .\n", vehicle.ID, err)
		return Record{}
	}
	return r
}

func (a *App) InsertNewVehicle(vehicle Vehicle) error {
	stmt := `INSERT INTO vehicles (
	plate, maker, model, year, assigned_to,
	location, photo_url) VALUES(?, ?, ?, ?, ?, ?, ?);`

	_, err := a.db.Exec(stmt, vehicle.Plate, vehicle.Maker,
		vehicle.Model, vehicle.Year, vehicle.AssignedTo, vehicle.Location, vehicle.PhotoURL)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return errors.New("El Técnico seleccionado ya tiene un vehículo asignado")
		}
		return err
	}
	return nil
}
