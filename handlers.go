package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ErrorPageData struct {
	Code    int
	Title   string
	Message string
}

// RenderError renders error.html with the right status code. If the template
// itself fails to render, it falls back to http.Error instead of recursing.
func (a *App) RenderError(w http.ResponseWriter, status int, title, message string) {
	var buf bytes.Buffer
	data := ErrorPageData{Code: status, Title: title, Message: message}
	if err := a.tmpl.ExecuteTemplate(&buf, "error.html", data); err != nil {
		log.Println("Error rendering error.html:", err)
		http.Error(w, message, status)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (a *App) NewVehicleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Render to a buffer first: if ExecuteTemplate fails partway through,
		// nothing has been written to w yet, so RenderError can still set the
		// status code correctly instead of appending onto a half-sent 200 response.
		var buf bytes.Buffer
		if err := a.tmpl.ExecuteTemplate(&buf, "new_vehicle.html", nil); err != nil {
			log.Println("Error rendering new_vehicle.html:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al cargar la página. Intenta de nuevo.")
			return
		}
		buf.WriteTo(w)
	}

	if r.Method == "POST" {
		// Add vehicle to db
		// TODO
		// Maybe parse returns a sequence of bytes?
		r.Body = http.MaxBytesReader(w, r.Body, 20<<20)
		maxMemory := int64(10 << 20)
		if parse_err := r.ParseMultipartForm(maxMemory); parse_err != nil {
			log.Println("Could not parse form values:", parse_err)
			a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", "No se pudo procesar el formulario enviado. Intenta de nuevo.")
			return
		}
		new_vehicle := NewVehicleFromForm(r)
		if insert_err := a.InsertVehicle(new_vehicle); insert_err != nil {
			log.Println("Error inserting new vehicle to database:", insert_err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al insertar vehículo nuevo en base de datos.")
			return
		}

		// Vehicle added with no errors. Now redirect
		http.Redirect(w, r, "/vehiculos/", http.StatusSeeOther)
		return
	}
}

func (a *App) HomeHandler(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/vehiculos/")

	// If there's no prefix, only 'vehiculos' was called, so we only need the index list
	if id == "" {
		// Get all vehicles from database
		vehicles, err := a.GetVehicles()
		if err != nil {
			// We don't have cars?
			// Was one car malformed or incomplete?
			log.Fatal("Error when querying for all cars: ", err)
		}
		// Execute template with cars
		err = a.tmpl.ExecuteTemplate(w, "index.html", vehicles)
		if err != nil {
			log.Fatal(err)
		}
		return
	} else {
		// User clicked on a car, so we need to obtain a specific car information
		// And call the vehicle template with that specific car info
		id, err := strconv.Atoi(id)
		if err != nil {
			fmt.Fprintf(w, "Error when converting id to int: %s\n", err)
			return
		}

		vehicle, err := a.GetVehicleByID(id)
		if err != nil {
			log.Fatal("Could not get vehicle from database: ", err)
		}
		err = a.tmpl.ExecuteTemplate(w, "vehicle.html", vehicle)
		if err != nil {
			fmt.Printf("Could not execute tempalte %s\n", err)
			return
		}
	}
}
