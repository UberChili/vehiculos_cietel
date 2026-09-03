package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const uploadDir = "uploads"

var allowedPhotoExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

// ValidationError marks a user-input problem (as opposed to a DB/server
// failure), so handlers can respond with 400 and the message as-is.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type Record struct {
	ID          int
	DateShort   string
	Type        string
	Description string
	Cost        string
}

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
	Records     []Record
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

func NewRecordFromForm(r *http.Request) Record {
	date_short := r.FormValue("date")
	record_type := r.FormValue("type")
	description := r.FormValue("description")
	cost := r.FormValue("cost")

	return Record{DateShort: date_short, Type: record_type, Description: description, Cost: cost}
}

// Same idea as Validate for the Vehicle type
func (r Record) Validate() error {
	if strings.TrimSpace(r.DateShort) == "" {
		return &ValidationError{"La fecha es obligatoria"}
	}
	// <input type="date"> always submits ISO format (YYYY-MM-DD) regardless
	// of locale, and time.Parse rejects out-of-range days/months (e.g. Feb 30)
	// on its own.
	if _, err := time.Parse("2006-01-02", r.DateShort); err != nil {
		return &ValidationError{"La fecha no es válida"}
	}
	if strings.TrimSpace(r.Type) == "" {
		return &ValidationError{"Seleccionar un tipo de reparación es obligatorio"}
	}
	if strings.TrimSpace(r.Description) == "" {
		return &ValidationError{"Ingresar una descripción de la reparación o mantenimiento es obligatorio"}
	}
	return nil
}

// Validate checks the fields required to save a vehicle. Client-side
// "required" on the form is not enough, since requests don't have to go
// through the browser.
func (v Vehicle) Validate() error {
	if strings.TrimSpace(v.Plate) == "" {
		return &ValidationError{"La placa es obligatoria"}
	}
	if strings.TrimSpace(v.Maker) == "" {
		return &ValidationError{"La marca es obligatoria"}
	}
	if strings.TrimSpace(v.Model) == "" {
		return &ValidationError{"El modelo es obligatorio"}
	}
	if strings.TrimSpace(v.Year) == "" {
		return &ValidationError{"El año es obligatorio"}
	}
	return nil
}

// SavePhoto reads the optional "photo" file field and stores it under
// uploadDir, returning the URL path to serve it from. Returns an empty
// string, nil error if no photo was submitted.
func SavePhoto(r *http.Request) (string, error) {
	file, header, err := r.FormFile("photo")
	if err == http.ErrMissingFile {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedPhotoExts[ext] {
		return "", &ValidationError{"Formato de imagen no soportado (usa jpg, png, gif o webp)"}
	}

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst, err := os.Create(filepath.Join(uploadDir, filename))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return "/" + uploadDir + "/" + filename, nil
}

// deleteUploadedPhoto removes a previously saved photo from disk when it's
// being replaced or cleared. Failures are only logged, not returned: a
// leftover file is harmless, and shouldn't block the vehicle update that
// triggered it.
func deleteUploadedPhoto(photoURL string) {
	prefix := "/" + uploadDir + "/"
	if photoURL == "" || !strings.HasPrefix(photoURL, prefix) {
		return
	}
	path := filepath.Join(uploadDir, strings.TrimPrefix(photoURL, prefix))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		log.Println("Could not delete old vehicle photo:", err)
	}
}

// last_service/last_repair are derived from the records table (most recent
// date per record type) rather than read from the vehicles.last_service
// column, which nothing ever writes to. Dates are stored as ISO 8601
// (YYYY-MM-DD), so MAX() on the text column sorts chronologically.
const vehicleSelect = `
	SELECT v.id, v.plate, v.maker, v.model, v.year, v.assigned_to, v.location, v.photo_url,
		COALESCE((SELECT MAX(date) FROM records WHERE vehicle_id = v.id AND type = 'Servicio'), '') AS last_service,
		COALESCE((SELECT MAX(date) FROM records WHERE vehicle_id = v.id AND type = 'Reparación'), '') AS last_repair
	FROM vehicles v`

func scanVehicle(row interface{ Scan(...any) error }, v *Vehicle) error {
	return row.Scan(&v.ID, &v.Plate, &v.Maker, &v.Model, &v.Year, &v.AssignedTo,
		&v.Location, &v.PhotoURL, &v.LastService, &v.LastRepair)
}

func (a *App) GetVehicleByID(id int) (Vehicle, error) {
	var v Vehicle
	row := a.db.QueryRow(vehicleSelect+" WHERE v.id = ?", id)
	if err := scanVehicle(row, &v); err != nil {
		return v, err
	}
	return v, nil
}

func (a *App) GetVehicleRecordsByID(id int) ([]Record, error) {
	rows, err := a.db.Query("SELECT id, date, type, description, cost FROM records WHERE vehicle_id = ? ORDER BY date DESC, id DESC", id)
	if err != nil {
		return nil, err
	}

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.ID, &r.DateShort, &r.Type, &r.Description, &r.Cost); err != nil {
			return records, err
		}
		records = append(records, r)
	}
	if err = rows.Err(); err != nil {
		return records, err
	}

	return records, nil
}

// GetRecordByID scopes the lookup to vehicle_id so a record can't be viewed
// through a URL for a vehicle it doesn't belong to.
func (a *App) GetRecordByID(vehicleID, recordID int) (Record, error) {
	var r Record
	row := a.db.QueryRow("SELECT id, date, type, description, cost FROM records WHERE id = ? AND vehicle_id = ?", recordID, vehicleID)
	err := row.Scan(&r.ID, &r.DateShort, &r.Type, &r.Description, &r.Cost)
	if err != nil {
		return r, err
	}
	return r, nil
}

func (a *App) GetVehicles() ([]Vehicle, error) {
	rows, err := a.db.Query(vehicleSelect)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []Vehicle

	// Loop through rows, using Scan to assign column data to struct fields
	for rows.Next() {
		var v Vehicle
		if err := scanVehicle(rows, &v); err != nil {
			return vehicles, err
		}
		vehicles = append(vehicles, v)
	}
	if err = rows.Err(); err != nil {
		return vehicles, err
	}

	return vehicles, nil
}
