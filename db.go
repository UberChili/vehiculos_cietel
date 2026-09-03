package main

import (
	"database/sql"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// CreateOrOpenTable opens (creating if needed) the SQLite database at path,
// and applies the schema/migrations. path is a plain filesystem path (or
// ":memory:"); the "?_foreign_keys=on" DSN parameter is appended here.
func CreateOrOpenTable(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
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

	// Migration for dbs created before last_service was derived from
	// records instead of stored directly on the vehicle row.
	if _, err := db.Exec("ALTER TABLE vehicles DROP COLUMN last_service"); err != nil &&
		!strings.Contains(err.Error(), "no such column") {
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
	_, err := a.db.Exec("INSERT INTO vehicles(plate, maker, model, year, assigned_to, location, photo_url) VALUES(?, ?, ?, ?, ?, ?, ?)",
		vehicle.Plate, vehicle.Maker, vehicle.Model, vehicle.Year, vehicle.AssignedTo, vehicle.Location, vehicle.PhotoURL)
	if err != nil {
		return err
	}
	log.Println("New vehicle inserted successfully")

	return nil
}

func (a *App) UpdateVehicle(vehicle Vehicle) error {
	if err := vehicle.Validate(); err != nil {
		return err
	}
	if len(vehicle.Plate) > 10 {
		return &ValidationError{"La placa es demasiado larga (máximo 10 caracteres)"}
	}
	_, err := a.db.Exec("UPDATE vehicles SET plate = ?, maker = ?, model = ?, year = ?, assigned_to = ?, location = ?, photo_url = ? WHERE id = ?",
		vehicle.Plate, vehicle.Maker, vehicle.Model, vehicle.Year, vehicle.AssignedTo, vehicle.Location, vehicle.PhotoURL, vehicle.ID)
	if err != nil {
		return err
	}
	log.Printf("Vehicle %d updated successfully\n", vehicle.ID)

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

// DeleteRecord removes a record, scoped to vehicle_id so a request can't
// delete a record belonging to a different vehicle.
func (a *App) DeleteRecord(vehicleID, recordID int) error {
	_, err := a.db.Exec("DELETE FROM records WHERE id = ? AND vehicle_id = ?", recordID, vehicleID)
	if err != nil {
		return err
	}
	log.Printf("Record %d deleted for vehicle %d\n", recordID, vehicleID)
	return nil
}
