package main

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "vehicles.db")
	if err != nil {
		log.Fatal("Could not open database: ", err)
	}
	log.Println("Successfully connected to SQLite database.")

	return db
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

// Checks is a vehicle is already assigned to a worker
// This would mean that another vehicle can not be assigned to the same worker
func (a *App) IsAssigned(technician_name string) bool {

}
