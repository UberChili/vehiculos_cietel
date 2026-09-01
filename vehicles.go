package main

import (
	"fmt"
	"io"
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

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
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

func (a *App) GetVehicleByID(id int) (Vehicle, error) {
	var v Vehicle
	row := a.db.QueryRow("SELECT id, plate, maker, model, year, assigned_to, location, last_service, photo_url FROM vehicles WHERE id = ?", id)
	err := row.Scan(&v.ID, &v.Plate, &v.Maker, &v.Model, &v.Year, &v.AssignedTo,
		&v.Location, &v.LastService, &v.PhotoURL)
	if err != nil {
		return v, err
	}
	return v, nil
}

func (a *App) GetVehicles() ([]Vehicle, error) {
	rows, err := a.db.Query("SELECT id, plate, maker, model, year, assigned_to, location, last_service, photo_url FROM vehicles")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []Vehicle

	// Loop through rows, using Scan to assign column data to struct fields
	for rows.Next() {
		var v Vehicle
		if err := rows.Scan(&v.ID, &v.Plate, &v.Maker, &v.Model, &v.Year, &v.AssignedTo,
			&v.Location, &v.LastService, &v.PhotoURL); err != nil {
			return vehicles, err
		}
		vehicles = append(vehicles, v)
	}
	if err = rows.Err(); err != nil {
		return vehicles, err
	}

	return vehicles, nil
}
