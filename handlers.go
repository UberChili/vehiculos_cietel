package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
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

func (a *App) IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Get all vehicles from database
	// Eventually, get only a number (Not sure how to handle this)
	vehicles, err := a.findAllVehicles()
	if err != nil {
		fmt.Println("Error getting vehicles:", err)
		return
	}

	// Render template
	var buf bytes.Buffer
	err = a.tmpl.ExecuteTemplate(&buf, "index.html", vehicles)
	if err != nil {
		log.Println("Error rendering index.html:", err)
		a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "Ocurrió un error al cargar la página.")
		return
	}
	fmt.Fprintf(w, "%s", buf.Bytes())
}

func (a *App) VehicleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fmt.Println("GET method called on VehicleHandler with id:",
			chi.URLParam(r, "id"))

		vehicle, v_err := a.GetVehicle(chi.URLParam(r, "id"))
		if v_err != nil {
			log.Printf("Error getting vehicle with id: %s: %s\n", chi.URLParam(r, "id"), v_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "No se encontró el vehículo.")
			return
		}

		// capitalize values for pretty printing
		vehicle.Maker = CapitalizeFirst(vehicle.Maker)
		vehicle.Model = CapitalizeFirst(vehicle.Model)
		vehicle.Location = CapitalizeFirst(vehicle.Location)

		// We first try executing the template and outputting to a buffer
		// Don't remember why but this is safer than trying to send directly to the writer
		var buf bytes.Buffer
		err := a.tmpl.ExecuteTemplate(&buf, "vehicle.html", vehicle)
		if err != nil {
			log.Println("Error rendering vehicle.html: ", err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}
		// Everything loaded correctly, we can output the template to actual output
		fmt.Fprintf(w, "%s", buf.Bytes())
	}
}

func (a App) NewVehicleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// We might need to get the technicians list here
		technicians, err := a.findAllTechnicians()
		if err != nil {
			log.Println("Error loading template:", err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}

		var buf bytes.Buffer
		err = a.tmpl.ExecuteTemplate(&buf, "new_vehicle.html", struct{ Technicians []Technician }{technicians})
		if err != nil {
			log.Println("Error loading template:", err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}

		fmt.Fprintf(w, "%s", buf.Bytes())
	}

	if r.Method == http.MethodPost {
		plate := strings.ToUpper(r.FormValue("plate"))
		maker := strings.ToLower(r.FormValue("maker"))
		model := strings.ToLower(r.FormValue("model"))
		year := r.FormValue("year")
		assigned_to := r.FormValue("assigned_to")
		location := strings.ToLower(r.FormValue("location"))
		photo := r.FormValue("photo")

		vehicle := Vehicle{}
		// assigned_to_int := 0

		// convert technician id to int
		if assigned_to != "" {
			assigned_to_int, conv_err := strconv.Atoi(assigned_to)
			if conv_err != nil {
				log.Println("Error in Form values. Invalid technician id conversion: ", conv_err)
				message := fmt.Sprintf("Valores de vehículo inválidos. ID de técnico inválido: %s\n", conv_err)
				a.RenderError(w, http.StatusBadRequest, "Error del servidor.", message)
				return
			}

			vehicle = Vehicle{Maker: maker, Model: model, Plate: plate, Year: year, AssignedTo: &assigned_to_int, Location: location, PhotoURL: photo}
		} else {
			// Not assigned to any technician
			vehicle = Vehicle{Maker: maker, Model: model, Plate: plate, Year: year, AssignedTo: nil, Location: location, PhotoURL: photo}
		}

		// Validate fields
		vehicle_validation_err := ValidateVehicleFields(vehicle)
		if vehicle_validation_err != nil {
			log.Println("Error in Form values. Invalid vehicle fields: ", vehicle_validation_err)
			message := fmt.Sprintf("Valores de vehículo inválidos: %s\n", vehicle_validation_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", message)
			return
		}
		// Insert
		insert_err := a.InsertNewVehicle(vehicle)
		if insert_err != nil {
			log.Println("Error Inserting vehicle to table: ", insert_err)
			message := fmt.Sprintf("Error al agregar vehículo: %s", insert_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", message)
			return
		}
		// return to main page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (a App) RecordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		log.Println("GET called on RecordHandler")

		vehicle, v_err := a.GetVehicle(chi.URLParam(r, "id"))
		if v_err != nil {
			log.Printf("Error getting vehicle with id: %s: %s\n", chi.URLParam(r, "id"), v_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "No se encontró el vehículo.")
			return
		}

		record, err := a.GetRecord(chi.URLParam(r, "id"), chi.URLParam(r, "record_id"))
		if err != nil {
			log.Println("Error getting record:", err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "Error al obtener reparación")
			return
		}

		// GetVehicleRecordsByID already orders latest-first; just drop the one
		// being viewed to build the "other records" list.
		allRecords, err := a.GetVehicleRecordsByID(chi.URLParam(r, "id"))
		if err != nil {
			log.Println("Error when querying for records:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al consultar los registros de vehículo.")
			return
		}

		var otherRecords []Record
		for _, other := range allRecords {
			if other.ID != record.ID {
				otherRecords = append(otherRecords, other)
			}
		}

		data := struct {
			Vehicle      Vehicle
			Record       Record
			OtherRecords []Record
		}{Vehicle: vehicle, Record: record, OtherRecords: otherRecords}

		// Render and print
		var buf bytes.Buffer
		err = a.tmpl.ExecuteTemplate(&buf, "record.html", data)
		if err != nil {
			log.Println("Error loading template:", err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}
		fmt.Fprintf(w, "%s", buf.Bytes())
	}
}

func (a App) NewRecordHandler(w http.ResponseWriter, r *http.Request) {
	// Called a GET method o
	if r.Method == http.MethodGet {
		log.Println("GET called on NewRecordHandler")

		vehicle_id, conv_err := strconv.Atoi(chi.URLParam(r, "id"))
		if conv_err != nil {
			log.Println("Error with vehicle id:", conv_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "Error con ID de vehículo. ID inválido.")
			return
		}

		// Display form for new record
		var buf bytes.Buffer
		err := a.tmpl.ExecuteTemplate(&buf, "new_record.html", Record{VehicleID: vehicle_id})
		if err != nil {
			log.Println("Error loading template:", err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}
		fmt.Fprintf(w, "%s", buf.Bytes())
	}

	if r.Method == http.MethodPost {
		log.Println("POST called on NewRecordHandler")

		// process data to add new record
		record_type := r.FormValue("type")
		vehicle_id, conv_err := strconv.Atoi(chi.URLParam(r, "id"))
		if conv_err != nil {
			log.Println("Error with vehicle id:", conv_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "Error con ID de vehículo. ID inválido.")
			return
		}

		date := r.FormValue("date")
		description := r.FormValue("description")
		cost := r.FormValue("cost")

		record := Record{VehicleID: vehicle_id, DateShort: date, Type: record_type, Description: description, Cost: cost}
		log.Println(record)

		insert_err := a.InsertNewRecord(record)
		if insert_err != nil {
			log.Println("Error inserting vehicle:", insert_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "Error al insertar reparación: ")
			return
		}
		// return to vehicle page
		// r.Get("/vehiculos/{id}/nuevo-registro", app.NewRecordHandler)
		vehicle_page_url := fmt.Sprintf("/vehiculos/%s", chi.URLParam(r, "id"))
		http.Redirect(w, r, vehicle_page_url, http.StatusSeeOther)
	}
}
