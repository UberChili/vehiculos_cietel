package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// SavePhoto stores the "photo" file from a multipart form in uploads/ and
// returns its URL (e.g. /uploads/1788294772393192185.jpg). The photo is
// optional, so when none was submitted it returns "" and no error.
func SavePhoto(r *http.Request) (string, error) {
	file, header, err := r.FormFile("photo")
	if err == http.ErrMissingFile {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer file.Close()

	if header.Size > 20<<20 {
		return "", errors.New("La foto pesa más de 20MB.")
	}

	// Look at the file's actual bytes (not its name) to make sure it's an
	// image a browser can display, and pick the extension from that.
	head := make([]byte, 512)
	n, _ := file.Read(head)
	extensions := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp", "image/gif": ".gif"}
	ext, ok := extensions[http.DetectContentType(head[:n])]
	if !ok {
		return "", errors.New("El archivo no es una imagen válida (JPG, PNG, WEBP o GIF).")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	if err := os.MkdirAll("uploads", 0755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst, err := os.Create(filepath.Join("uploads", name))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}

// ServiceIntervalDays is how often a vehicle needs a service.
// PLACEHOLDER: the real interval is still to be decided. It's only here so the
// status badges on the index page have something to work with.
const ServiceIntervalDays = 90

// ServiceDueSoonDays is how many days before its due date a vehicle shows as "Próximo"
const ServiceDueSoonDays = 15

// ServiceStatus tells how a vehicle is doing on services, from its last service
// date (ISO, as stored): "ok", "warning" (due soon), "overdue" or "none" (never
// serviced). The values match the status-* CSS classes of the badges in index.html.
func ServiceStatus(lastService string) string {
	last, err := time.Parse("2006-01-02", lastService)
	if err != nil {
		return "none"
	}
	due := last.AddDate(0, 0, ServiceIntervalDays)
	switch {
	case time.Now().After(due):
		return "overdue"
	case time.Now().After(due.AddDate(0, 0, -ServiceDueSoonDays)):
		return "warning"
	default:
		return "ok"
	}
}

// FormatCost shows a cost stored in cents as pesos, e.g. 185050 -> "1850.50"
func FormatCost(cents *int64) string {
	if cents == nil {
		return ""
	}
	return fmt.Sprintf("%d.%02d", *cents/100, *cents%100)
}

// ParseRecordForm reads and validates the fields shared by the new and edit
// record forms. VehicleID and ID are left for the caller to set.
func ParseRecordForm(r *http.Request) (Record, error) {
	record := Record{Type: r.FormValue("type"), Description: strings.TrimSpace(r.FormValue("description"))}

	// Last service/repair are looked up by these exact type names
	if record.Type != "Servicio" && record.Type != "Reparación" {
		return Record{}, errors.New("El tipo debe ser Servicio o Reparación.")
	}
	date, err := time.Parse("02-01-2006", r.FormValue("date"))
	if err != nil {
		return Record{}, errors.New("Fecha inválida. Usa el formato dd-mm-aaaa.")
	}
	record.DateShort = date.Format("2006-01-02")
	if record.Description == "" {
		return Record{}, errors.New("La descripción es obligatoria.")
	}

	// Cost is optional. Stored in cents so money never has float rounding errors
	if cost := strings.TrimSpace(r.FormValue("cost")); cost != "" {
		pesos, err := strconv.ParseFloat(cost, 64)
		// written this way so it also rejects NaN and Inf, which ParseFloat accepts
		if err != nil || !(pesos >= 0 && pesos < 1_000_000_000) {
			return Record{}, errors.New("Costo inválido.")
		}
		cents := int64(math.Round(pesos * 100))
		record.Cost = &cents
	}
	return record, nil
}

// NormalizeName lowercases a name and removes extra spaces, so " Juan" and "juan"
// don't end up as two different spellings. It's capitalized again for display.
func NormalizeName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(name), " "))
}

// FormatDateDisplay converts a date stored as ISO 8601 (YYYY-MM-DD) into
// dd-mm-yyyy, the format shown to users everywhere in the UI. Dates stay
// ISO in the database (so MAX()/ORDER BY sort correctly); this only
// affects display. Falls back to the original string unchanged if it
// isn't a valid ISO date — covers the empty "no service yet" case, and
// avoids ever rendering a blank date as something misleading.
func FormatDateDisplay(iso string) string {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	return t.Format("02-01-2006")
}

// IsAssignedTo reports whether a vehicle's AssignedTo points to the given
// technician ID. Templates can't compare a *int with an int using eq, so
// this does the nil check and dereference for them.
func IsAssignedTo(assignedTo *int, technicianID int) bool {
	return assignedTo != nil && *assignedTo == technicianID
}

// helper function to capitalize strings
func CapitalizeFirst(text string) string {
	if text == "" {
		return ""
	}
	runes := []rune(text)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// CapitalizeWords capitalizes every word, for names like "maría josé" -> "María José"
func CapitalizeWords(text string) string {
	words := strings.Fields(text)
	for i, word := range words {
		words[i] = CapitalizeFirst(word)
	}
	return strings.Join(words, " ")
}

func ValidateVehicleFields(vehicle Vehicle) error {
	if vehicle.Plate == "" || len(vehicle.Plate) >= 10 {
		return errors.New("Invalid Plate.")
	}
	models, ok := ModelsByMaker[vehicle.Maker]
	if !ok {
		return errors.New("Marca inválida. No en la lista de Marcas de vehículos.")
	}
	if !slices.Contains(models, vehicle.Model) {
		return errors.New("Modelo invalido. No es un modelo de esa marca.")
	}
	year, err := strconv.Atoi(vehicle.Year)
	if err != nil {
		return err
	}
	// Next year's models are already sold, so allow up to current year + 1
	if year > time.Now().Year()+1 || year <= 2009 {
		return errors.New("Invalid Year.")
	}
	// Location is optional, but when set it must be "City, State" from CitiesByState
	if vehicle.Location != "" {
		city, state, _ := strings.Cut(vehicle.Location, ", ")
		if !slices.Contains(CitiesByState[state], city) {
			return errors.New("Ubicación inválida. No en la lista de ubicaciones.")
		}
	}

	return nil
}
