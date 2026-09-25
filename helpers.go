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

// How often a vehicle needs a general service and a change of bandas.
// NOT FINAL: these are what the cousin suggested on 2026-09-24 (service every
// 8 months, bandas every 6), still to be agreed on as a family.
const ServiceIntervalMonths = 8
const BandasIntervalMonths = 6

// DueSoonDays is how many days before its due date a vehicle shows as "Próximo"
const DueSoonDays = 15

// ServiceStatus tells how a vehicle is doing on general services, from its last
// service date (ISO, as stored). See dueStatus for the values.
func ServiceStatus(lastService string) string {
	return dueStatus(lastService, ServiceIntervalMonths)
}

// BandasStatus is the same for the change of bandas, from the last record marked
// "Se cambiaron las bandas" (Vehicle.LastBandas).
func BandasStatus(lastBandas string) string {
	return dueStatus(lastBandas, BandasIntervalMonths)
}

// dueStatus returns "ok", "warning" (due soon), "overdue" or "none" (never done)
// for something that has to be done every intervalMonths. The values match the
// status-* CSS classes of the badges in index.html.
func dueStatus(lastDate string, intervalMonths int) string {
	last, err := time.Parse("2006-01-02", lastDate)
	if err != nil {
		return "none"
	}
	due := last.AddDate(0, intervalMonths, 0)
	switch {
	case time.Now().After(due):
		return "overdue"
	case time.Now().After(due.AddDate(0, 0, -DueSoonDays)):
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

	// Odometer is optional: several vehicles have a broken one. No reading is
	// stored as nil (NULL), never as 0
	record.OdometerBroken = r.FormValue("odometer_broken") != ""
	record.BandasChanged = r.FormValue("bandas_changed") != ""
	if km := strings.TrimSpace(r.FormValue("odometer_km")); km != "" && !record.OdometerBroken {
		n, err := strconv.Atoi(km)
		if err != nil || n < 0 || n > 9_999_999 {
			return Record{}, errors.New("Kilometraje inválido.")
		}
		record.OdometerKm = &n
	}

	if days := strings.TrimSpace(r.FormValue("downtime_days")); days != "" {
		n, err := strconv.Atoi(days)
		if err != nil || n < 0 || n > 365 {
			return Record{}, errors.New("Días sin poder usarse inválidos.")
		}
		record.DowntimeDays = &n
	}

	// Cause only applies to repairs, and there it's required ("No se sabe" is the way out)
	if record.Type == "Reparación" {
		record.Cause = r.FormValue("cause")
		if !slices.Contains(RepairCauses, record.Cause) {
			return Record{}, errors.New("Selecciona la causa de la reparación.")
		}
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
	if year > time.Now().Year()+1 || year <= 2000 {
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
