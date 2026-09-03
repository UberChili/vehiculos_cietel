package main

import (
	"bytes"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

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
		r.Body = http.MaxBytesReader(w, r.Body, 20<<20)
		maxMemory := int64(10 << 20)
		if parse_err := r.ParseMultipartForm(maxMemory); parse_err != nil {
			log.Println("Could not parse form values:", parse_err)
			a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", "No se pudo procesar el formulario enviado. Intenta de nuevo.")
			return
		}
		photo_url, photo_err := SavePhoto(r)
		if photo_err != nil {
			var verr *ValidationError
			if errors.As(photo_err, &verr) {
				a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", verr.Message)
				return
			}
			log.Println("Error saving vehicle photo:", photo_err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al guardar la foto.")
			return
		}

		new_vehicle := NewVehicleFromForm(r)
		new_vehicle.PhotoURL = photo_url
		if insert_err := a.InsertVehicle(new_vehicle); insert_err != nil {
			var verr *ValidationError
			if errors.As(insert_err, &verr) {
				a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", verr.Message)
				return
			}
			log.Println("Error inserting new vehicle to database:", insert_err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al insertar vehículo nuevo en base de datos.")
			return
		}

		// Vehicle added with no errors. Now redirect
		http.Redirect(w, r, "/vehiculos/", http.StatusSeeOther)
		return
	}
}

func (a *App) NewRecordHandler(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	if id == "" {
		log.Println("No id in URL?")
		a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "No se especificó un vehículo.")
		return
	}

	if req.Method == "GET" {
		// Render to a buffer first: if ExecuteTemplate fails partway through,
		// nothing has been written to w yet, so RenderError can still set the
		// status code correctly instead of appending onto a half-sent 200 response.
		var buf bytes.Buffer
		data := struct{ VehicleID string }{VehicleID: id}
		if err := a.tmpl.ExecuteTemplate(&buf, "new_record.html", data); err != nil {
			log.Println("Error rendering new_record.html:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al cargar la página. Intenta de nuevo.")
			return
		}
		buf.WriteTo(w)
	}

	if req.Method == "POST" {
		if parse_err := req.ParseForm(); parse_err != nil {
			log.Println("Could not parse form values:", parse_err)
			a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", "No se pudo procesar el formulario enviado. Intenta de nuevo.")
			return
		}
		new_record := NewRecordFromForm(req)
		if insert_err := a.InsertRecord(id, new_record); insert_err != nil {
			var verr *ValidationError
			if errors.As(insert_err, &verr) {
				a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", verr.Message)
				return
			}
			log.Println("Error inserting new register to database:", insert_err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al insertar registro de reparación nuevo en base de datos para vehículo.")
			return
		}

		// Register added with no errors. Now redirect
		http.Redirect(w, req, "/vehiculos/", http.StatusSeeOther)
		return
	}
}

func (a *App) RecordHandler(w http.ResponseWriter, req *http.Request) {
	vehicleIDParam := chi.URLParam(req, "id")
	vehicleID, err := strconv.Atoi(vehicleIDParam)
	if err != nil {
		a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", "El identificador del vehículo no es válido.")
		return
	}

	recordIDParam := chi.URLParam(req, "record_id")
	recordID, err := strconv.Atoi(recordIDParam)
	if err != nil {
		a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", "El identificador del registro no es válido.")
		return
	}

	vehicle, err := a.GetVehicleByID(vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		a.RenderError(w, http.StatusNotFound, "No encontrado", "No existe un vehículo con ese identificador.")
		return
	}
	if err != nil {
		log.Println("Could not get vehicle from database:", err)
		a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al consultar el vehículo.")
		return
	}

	record, err := a.GetRecordByID(vehicleID, recordID)
	if errors.Is(err, sql.ErrNoRows) {
		a.RenderError(w, http.StatusNotFound, "No encontrado", "No existe un registro con ese identificador para este vehículo.")
		return
	}
	if err != nil {
		log.Println("Could not get record from database:", err)
		a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al consultar el registro.")
		return
	}

	// GetVehicleRecordsByID already orders latest-first; just drop the one
	// being viewed to build the "other records" list.
	allRecords, err := a.GetVehicleRecordsByID(vehicleID)
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

	var buf bytes.Buffer
	if err := a.tmpl.ExecuteTemplate(&buf, "record.html", data); err != nil {
		log.Println("Error rendering record.html:", err)
		a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al cargar la página.")
		return
	}
	buf.WriteTo(w)
}

func (a *App) DeleteRecordHandler(w http.ResponseWriter, req *http.Request) {
	vehicleIDParam := chi.URLParam(req, "id")
	vehicleID, err := strconv.Atoi(vehicleIDParam)
	if err != nil {
		a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", "El identificador del vehículo no es válido.")
		return
	}

	recordIDParam := chi.URLParam(req, "record_id")
	recordID, err := strconv.Atoi(recordIDParam)
	if err != nil {
		a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", "El identificador del registro no es válido.")
		return
	}

	if err := a.DeleteRecord(vehicleID, recordID); err != nil {
		log.Println("Error deleting record from database:", err)
		a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al eliminar el registro.")
		return
	}

	http.Redirect(w, req, "/vehiculos/"+vehicleIDParam+"/", http.StatusSeeOther)
}

func (a *App) EditVehicleHandler(w http.ResponseWriter, req *http.Request) {
	idParam := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", "El identificador del vehículo no es válido.")
		return
	}

	vehicle, err := a.GetVehicleByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		a.RenderError(w, http.StatusNotFound, "No encontrado", "No existe un vehículo con ese identificador.")
		return
	}
	if err != nil {
		log.Println("Could not get vehicle from database:", err)
		a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al consultar el vehículo.")
		return
	}

	if req.Method == "GET" {
		var buf bytes.Buffer
		if err := a.tmpl.ExecuteTemplate(&buf, "edit_vehicle.html", vehicle); err != nil {
			log.Println("Error rendering edit_vehicle.html:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al cargar la página. Intenta de nuevo.")
			return
		}
		buf.WriteTo(w)
		return
	}

	if req.Method == "POST" {
		req.Body = http.MaxBytesReader(w, req.Body, 20<<20)
		maxMemory := int64(10 << 20)
		if parse_err := req.ParseMultipartForm(maxMemory); parse_err != nil {
			log.Println("Could not parse form values:", parse_err)
			a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", "No se pudo procesar el formulario enviado. Intenta de nuevo.")
			return
		}

		new_photo_url, photo_err := SavePhoto(req)
		if photo_err != nil {
			var verr *ValidationError
			if errors.As(photo_err, &verr) {
				a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", verr.Message)
				return
			}
			log.Println("Error saving vehicle photo:", photo_err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al guardar la foto.")
			return
		}

		updated_vehicle := NewVehicleFromForm(req)
		updated_vehicle.ID = id

		// A newly uploaded photo always wins; otherwise the "delete_photo"
		// checkbox clears it; otherwise the existing photo is kept as-is.
		delete_photo := req.PostFormValue("delete_photo") == "on"
		switch {
		case new_photo_url != "":
			deleteUploadedPhoto(vehicle.PhotoURL)
			updated_vehicle.PhotoURL = new_photo_url
		case delete_photo:
			deleteUploadedPhoto(vehicle.PhotoURL)
			updated_vehicle.PhotoURL = ""
		default:
			updated_vehicle.PhotoURL = vehicle.PhotoURL
		}

		if update_err := a.UpdateVehicle(updated_vehicle); update_err != nil {
			var verr *ValidationError
			if errors.As(update_err, &verr) {
				a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", verr.Message)
				return
			}
			log.Println("Error updating vehicle in database:", update_err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al actualizar el vehículo.")
			return
		}

		http.Redirect(w, req, "/vehiculos/"+idParam+"/", http.StatusSeeOther)
		return
	}
}

func (a *App) HomeHandler(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	// If there's no prefix, only 'vehiculos' was called, so we only need the index list
	if id == "" {
		// Get all vehicles from database
		vehicles, err := a.GetVehicles()
		if err != nil {
			log.Println("Error when querying for all cars:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al consultar los vehículos.")
			return
		}

		// Execute template with cars
		if err := a.tmpl.ExecuteTemplate(w, "index.html", vehicles); err != nil {
			log.Println("Error rendering index.html:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al cargar la página.")
			return
		}
		return
	} else {
		// User clicked on a car, so we need to obtain a specific car information
		// And call the vehicle template with that specific car info
		id, err := strconv.Atoi(id)
		if err != nil {
			a.RenderError(w, http.StatusBadRequest, "Solicitud inválida", "El identificador del vehículo no es válido.")
			return
		}

		vehicle, err := a.GetVehicleByID(id)
		if errors.Is(err, sql.ErrNoRows) {
			a.RenderError(w, http.StatusNotFound, "No encontrado", "No existe un vehículo con ese identificador.")
			return
		}
		if err != nil {
			log.Println("Could not get vehicle from database:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al consultar el vehículo.")
			return
		}

		// Get all records from a vehicle
		records, err := a.GetVehicleRecordsByID(id)
		if err != nil {
			log.Println("Error when querying for records:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al consultar los registros de vehículo.")
			return
		}

		vehicle.Records = records

		if err := a.tmpl.ExecuteTemplate(w, "vehicle.html", vehicle); err != nil {
			log.Println("Error rendering vehicle.html:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor", "Ocurrió un error al cargar la página.")
			return
		}
	}
}
