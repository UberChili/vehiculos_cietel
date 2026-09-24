package main

import (
	"errors"
	"fmt"
	"io"
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

	return nil
}
