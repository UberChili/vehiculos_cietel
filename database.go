package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

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
	technicians_table_stmt := `CREATE TABLE IF NOT EXISTS technicians (id INTEGER PRIMARY KEY AUTOINCREMENT, first_name TEXT NOT NULL, last_name TEXT NOT NULL,
				active INTEGER NOT NULL DEFAULT 1);`
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
				cost INTEGER,
				odometer_km INTEGER,
				odometer_broken INTEGER NOT NULL DEFAULT 0,
				downtime_days INTEGER,
				cause TEXT,
				bandas_changed INTEGER NOT NULL DEFAULT 0,
				technician_id INTEGER REFERENCES technicians(id) ON DELETE SET NULL
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
	// Plates are unique, ignoring dashes and spaces ("DCE-456" = "DCE 456" = "DCE456")
	plate_index_stmt := `CREATE UNIQUE INDEX IF NOT EXISTS vehicles_plate_unique
				ON vehicles (replace(replace(plate, '-', ''), ' ', ''));`
	_, err = db.Exec(plate_index_stmt)
	if err != nil {
		return nil, fmt.Errorf("error creating plates index: %w", err)
	}

	log.Println("Succesfully opened tables.")
	return db, nil
}

func (a App) findAllVehicles() ([]Vehicle, error) {
	// query := `SELECT id, plate, maker, model, year, assigned_to, location FROM vehicles ORDER BY LOCATION`
	query := `SELECT v.id, v.plate, v.maker, v.model, v.year, v.assigned_to,
			COALESCE(t.first_name || ' ' || t.last_name, 'Sin asignar') AS assigned_to_name, v.location,
			COALESCE((SELECT MAX(date) FROM records
					WHERE vehicle_id = v.id AND type = 'Servicio'), '') AS last_service,
			COALESCE((SELECT MAX(date) FROM records
					WHERE vehicle_id = v.id AND bandas_changed = 1), '') AS last_bandas
			FROM vehicles v
			LEFT JOIN technicians t ON v.assigned_to = t.id
			ORDER BY v.location`

	rows, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []Vehicle

	for rows.Next() {
		v := &Vehicle{}
		err := rows.Scan(
			&v.ID, &v.Plate, &v.Maker, &v.Model, &v.Year, &v.AssignedTo, &v.AssignedToName, &v.Location, &v.LastService, &v.LastBandas)

		if err != nil {
			return nil, err
		}

		v.Maker = CapitalizeFirst(v.Maker)
		v.Model = CapitalizeFirst(v.Model)
		if v.AssignedTo != nil { // otherwise it's "Sin asignar"
			v.AssignedToName = CapitalizeWords(v.AssignedToName)
		}
		vehicles = append(vehicles, *v)
	}
	return vehicles, nil
}

func (a App) GetVehicle(id string) (Vehicle, error) {
	// row := a.db.QueryRow("SELECT id, plate, maker, model, year, assigned_to, location FROM vehicles WHERE id = ?", id)
	query := `SELECT v.id, v.plate, v.maker, v.model, v.year, v.assigned_to,
				COALESCE(t.first_name || ' ' || t.last_name, 'Sin asignar') AS asigned_to_name, location,
				COALESCE((SELECT MAX(date) FROM records
						WHERE vehicle_id = v.id AND type = 'Servicio'), '') AS last_service,
				COALESCE((SELECT MAX(date) FROM records
						WHERE vehicle_id = v.id AND type = 'Reparación'), '') AS last_repair,
				-- Any record (service or repair) marked "Se cambiaron las bandas"
				COALESCE((SELECT MAX(date) FROM records
						WHERE vehicle_id = v.id AND bandas_changed = 1), '') AS last_bandas,
				v.photo_url
				FROM vehicles v
				LEFT JOIN technicians t ON v.assigned_to = t.id
				WHERE v.id = ?`
	row := a.db.QueryRow(query, id)

	v := Vehicle{}

	err := row.Scan(&v.ID, &v.Plate, &v.Maker, &v.Model, &v.Year, &v.AssignedTo, &v.AssignedToName,
		&v.Location, &v.LastService, &v.LastRepair, &v.LastBandas, &v.PhotoURL)

	// Not neccessarily an error, but no results
	if err == sql.ErrNoRows {
		return Vehicle{}, errors.New("Vehicle not found")
	}
	if err != nil {
		return Vehicle{}, err
	}

	records_rows, err := a.db.Query("SELECT id, date, type, description, cost, bandas_changed FROM records WHERE vehicle_id = ? ORDER BY date DESC", id)
	if err != nil {
		v.Records = nil
		return v, err
	}

	for records_rows.Next() {
		r := &Record{}
		err := records_rows.Scan(&r.ID, &r.DateShort, &r.Type, &r.Description, &r.Cost, &r.BandasChanged)

		if err != nil {
			continue
		}
		v.Records = append(v.Records, *r)
	}

	return v, nil
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
			// Two UNIQUE rules on vehicles: the plate index and assigned_to
			if strings.Contains(err.Error(), "vehicles_plate_unique") {
				return errors.New("Ya existe un vehículo con esas placas")
			}
			return errors.New("El Técnico seleccionado ya tiene un vehículo asignado")
		}
		return err
	}
	return nil
}

// UpdateVehicle overwrites a vehicle's editable fields, including photo_url.
func (a *App) UpdateVehicle(vehicle Vehicle) error {
	stmt := `UPDATE vehicles SET plate = ?, maker = ?, model = ?, year = ?,
			assigned_to = ?, location = ?, photo_url = ? WHERE id = ?`

	_, err := a.db.Exec(stmt, vehicle.Plate, vehicle.Maker, vehicle.Model,
		vehicle.Year, vehicle.AssignedTo, vehicle.Location, vehicle.PhotoURL, vehicle.ID)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			// Two UNIQUE rules on vehicles: the plate index and assigned_to
			if strings.Contains(err.Error(), "vehicles_plate_unique") {
				return errors.New("Ya existe un vehículo con esas placas")
			}
			return errors.New("El Técnico seleccionado ya tiene un vehículo asignado")
		}
		return err
	}
	return nil
}

// DeleteVehicle removes a vehicle. Its records go with it (ON DELETE CASCADE).
func (a *App) DeleteVehicle(id string) error {
	_, err := a.db.Exec(`DELETE FROM vehicles WHERE id = ?`, id)
	return err
}

func (a *App) GetRecord(vehicle_id, record_id string) (Record, error) {
	query := `SELECT r.id, r.vehicle_id, r.date, r.type, r.description, r.cost,
				r.odometer_km, r.odometer_broken, r.downtime_days, COALESCE(r.cause, ''),
				r.bandas_changed, r.technician_id, COALESCE(t.first_name || ' ' || t.last_name, '')
				FROM records r
				LEFT JOIN technicians t ON r.technician_id = t.id
				WHERE r.id = ? AND r.vehicle_id = ?`
	row := a.db.QueryRow(query, record_id, vehicle_id)

	r := Record{}

	err := row.Scan(&r.ID, &r.VehicleID, &r.DateShort, &r.Type, &r.Description, &r.Cost,
		&r.OdometerKm, &r.OdometerBroken, &r.DowntimeDays, &r.Cause,
		&r.BandasChanged, &r.TechnicianID, &r.TechnicianName)

	// Not neccessarily an error, but no results
	if err == sql.ErrNoRows {
		return Record{}, errors.New("Record not found")
	}
	if err != nil {
		return Record{}, err
	}
	r.TechnicianName = CapitalizeWords(r.TechnicianName)

	return r, nil
}

func (a *App) GetVehicleRecordsByID(id string) ([]Record, error) {
	rows, err := a.db.Query("SELECT id, vehicle_id, date, type, description, cost FROM records WHERE vehicle_id = ? ORDER BY date DESC, id DESC", id)
	if err != nil {
		return nil, err
	}

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.ID, &r.VehicleID, &r.DateShort, &r.Type, &r.Description, &r.Cost); err != nil {
			return records, err
		}
		records = append(records, r)
	}
	if err = rows.Err(); err != nil {
		return records, err
	}

	return records, nil
}

func (a *App) InsertNewRecord(record Record) error {
	// technician_id is a snapshot of whoever has the vehicle right now, so the
	// record stays attributed to them even if the vehicle is reassigned later.
	// Only for recent records (see TechnicianSnapshotMaxDays): for old history
	// being entered today, whoever has the car now may not be who had it then,
	// so it's left NULL (unknown) instead of guessing.
	// NULLIF stores services' empty cause as NULL instead of ''.
	stmt := `INSERT INTO records (
			vehicle_id, date, type, description, cost,
			odometer_km, odometer_broken, downtime_days, cause, bandas_changed, technician_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?,
				CASE WHEN ? >= date('now', ?)
					THEN (SELECT assigned_to FROM vehicles WHERE id = ?) END)`

	max_age := fmt.Sprintf("-%d days", TechnicianSnapshotMaxDays)
	_, err := a.db.Exec(stmt, record.VehicleID, record.DateShort, record.Type, record.Description, record.Cost,
		record.OdometerKm, record.OdometerBroken, record.DowntimeDays, record.Cause, record.BandasChanged,
		record.DateShort, max_age, record.VehicleID)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) UpdateRecord(record Record) error {
	// technician_id is left as it was: it records who had the vehicle back then
	stmt := `UPDATE records SET date = ?, type = ?, description = ?, cost = ?,
			odometer_km = ?, odometer_broken = ?, downtime_days = ?, cause = NULLIF(?, ''),
			bandas_changed = ?
			WHERE id = ? AND vehicle_id = ?`

	_, err := a.db.Exec(stmt, record.DateShort, record.Type, record.Description, record.Cost,
		record.OdometerKm, record.OdometerBroken, record.DowntimeDays, record.Cause,
		record.BandasChanged, record.ID, record.VehicleID)
	return err
}

// GetOdometerReadings returns a vehicle's km readings, oldest first, leaving out
// one record (the one being edited; 0 leaves out none).
func (a *App) GetOdometerReadings(vehicle_id string, except_record_id int) ([]OdometerReading, error) {
	rows, err := a.db.Query(`SELECT date, odometer_km FROM records
			WHERE vehicle_id = ? AND id != ? AND odometer_km IS NOT NULL
			ORDER BY date`, vehicle_id, except_record_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []OdometerReading
	for rows.Next() {
		var reading OdometerReading
		if err := rows.Scan(&reading.Date, &reading.Km); err != nil {
			return nil, err
		}
		readings = append(readings, reading)
	}
	return readings, rows.Err()
}

func (a *App) DeleteRecord(vehicle_id, record_id string) (int64, error) {
	stmt := `DELETE FROM records WHERE id = ? AND vehicle_id = ?`

	result, err := a.db.Exec(stmt, record_id, vehicle_id)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
