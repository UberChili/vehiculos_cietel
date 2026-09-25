package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
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
		log.Println("Error getting vehicles:", err)
		a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar los vehículos.")
		return
	}

	// For the technicians view of the index page
	technicians, err := a.findAllTechnicians()
	if err != nil {
		log.Println("Error getting technicians:", err)
		a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar los técnicos.")
		return
	}

	data := struct {
		Vehicles    []Vehicle
		Technicians []Technician
	}{vehicles, technicians}

	// Render template
	var buf bytes.Buffer
	err = a.tmpl.ExecuteTemplate(&buf, "index.html", data)
	if err != nil {
		log.Println("Error rendering index.html:", err)
		a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar la página.")
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
		if vehicle.AssignedTo != nil { // otherwise it's "Sin asignar"
			vehicle.AssignedToName = CapitalizeWords(vehicle.AssignedToName)
		}

		// We first try executing the template and outputting to a buffer
		// Don't remember why but this is safer than trying to send directly to the writer
		var buf bytes.Buffer
		err := a.tmpl.ExecuteTemplate(&buf, "vehicle.html", vehicle)
		if err != nil {
			log.Println("Error rendering vehicle.html: ", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar la página.")
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
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}

		var buf bytes.Buffer
		data := struct {
			Technicians   []Technician
			ModelsByMaker map[string][]string
			CitiesByState map[string][]string
		}{technicians, ModelsByMaker, CitiesByState}

		err = a.tmpl.ExecuteTemplate(&buf, "new_vehicle.html", data)
		if err != nil {
			log.Println("Error loading template:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}

		fmt.Fprintf(w, "%s", buf.Bytes())
	}

	if r.Method == http.MethodPost {
		plate := strings.ToUpper(strings.TrimSpace(r.FormValue("plate")))
		maker := strings.ToLower(r.FormValue("maker"))
		model := strings.ToLower(r.FormValue("model"))
		year := r.FormValue("year")
		assigned_to := r.FormValue("assigned_to")
		// Comes from the CitiesByState dropdown, already written as it should be stored
		location := r.FormValue("location")
		photo, photo_err := SavePhoto(r)
		if photo_err != nil {
			log.Println("Error saving photo: ", photo_err)
			message := fmt.Sprintf("Error al guardar la foto: %s", photo_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", message)
			return
		}

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
		if assign_err := a.CheckTechnicianAssignable(vehicle.AssignedTo); assign_err != nil {
			a.RenderError(w, http.StatusBadRequest, "Técnico inválido.", assign_err.Error())
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

func (a App) EditVehicleHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if r.Method == http.MethodGet {
		vehicle, v_err := a.GetVehicle(id)
		if v_err != nil {
			log.Printf("Error getting vehicle with id: %s: %s\n", id, v_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "No se encontró el vehículo.")
			return
		}

		technicians, err := a.findAllTechnicians()
		if err != nil {
			log.Println("Error getting technicians:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}

		// edit_vehicle.html reads the vehicle fields, .Technicians, .ModelsByMaker and .CitiesByState at the same level
		data := struct {
			Vehicle
			Technicians   []Technician
			ModelsByMaker map[string][]string
			CitiesByState map[string][]string
		}{vehicle, technicians, ModelsByMaker, CitiesByState}

		var buf bytes.Buffer
		err = a.tmpl.ExecuteTemplate(&buf, "edit_vehicle.html", data)
		if err != nil {
			log.Println("Error loading template:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}
		fmt.Fprintf(w, "%s", buf.Bytes())
	}

	if r.Method == http.MethodPost {
		vehicle_id, conv_err := strconv.Atoi(id)
		if conv_err != nil {
			log.Println("Error converting vehicle id:", conv_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "ID de vehículo inválido.")
			return
		}

		vehicle := Vehicle{
			ID:       vehicle_id,
			Plate:    strings.ToUpper(strings.TrimSpace(r.FormValue("plate"))),
			Maker:    strings.ToLower(r.FormValue("maker")),
			Model:    strings.ToLower(r.FormValue("model")),
			Year:     r.FormValue("year"),
			Location: r.FormValue("location"), // from the CitiesByState dropdown
		}

		// Empty means "Sin asignar", so AssignedTo stays nil
		if assigned_to := r.FormValue("assigned_to"); assigned_to != "" {
			assigned_to_int, conv_err := strconv.Atoi(assigned_to)
			if conv_err != nil {
				log.Println("Error in Form values. Invalid technician id conversion: ", conv_err)
				a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "ID de técnico inválido.")
				return
			}
			vehicle.AssignedTo = &assigned_to_int
		}

		vehicle_validation_err := ValidateVehicleFields(vehicle)
		if vehicle_validation_err != nil {
			log.Println("Error in Form values. Invalid vehicle fields: ", vehicle_validation_err)
			message := fmt.Sprintf("Valores de vehículo inválidos: %s\n", vehicle_validation_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", message)
			return
		}
		if assign_err := a.CheckTechnicianAssignable(vehicle.AssignedTo); assign_err != nil {
			a.RenderError(w, http.StatusBadRequest, "Técnico inválido.", assign_err.Error())
			return
		}

		// Photo: keep the current one, unless it's deleted or a new one is uploaded
		old_vehicle, v_err := a.GetVehicle(id)
		if v_err != nil {
			log.Printf("Error getting vehicle with id: %s: %s\n", id, v_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "No se encontró el vehículo.")
			return
		}
		vehicle.PhotoURL = old_vehicle.PhotoURL
		if r.FormValue("delete_photo") != "" {
			vehicle.PhotoURL = ""
		}
		new_photo, photo_err := SavePhoto(r)
		if photo_err != nil {
			log.Println("Error saving photo: ", photo_err)
			message := fmt.Sprintf("Error al guardar la foto: %s", photo_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", message)
			return
		}
		if new_photo != "" {
			vehicle.PhotoURL = new_photo
		}

		update_err := a.UpdateVehicle(vehicle)
		if update_err != nil {
			log.Println("Error updating vehicle: ", update_err)
			message := fmt.Sprintf("Error al editar vehículo: %s", update_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", message)
			return
		}

		// The old photo was replaced or deleted, so remove its file from uploads/
		if old_vehicle.PhotoURL != "" && old_vehicle.PhotoURL != vehicle.PhotoURL {
			os.Remove(strings.TrimPrefix(old_vehicle.PhotoURL, "/"))
		}

		// back to the vehicle's page
		http.Redirect(w, r, "/vehiculos/"+id, http.StatusSeeOther)
	}
}

func (a App) DeleteVehicleHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Needed first to know which photo file to remove afterwards
	vehicle, v_err := a.GetVehicle(id)
	if v_err != nil {
		log.Printf("Error getting vehicle with id: %s: %s\n", id, v_err)
		a.RenderError(w, http.StatusNotFound, "Error del servidor.", "No se encontró el vehículo.")
		return
	}

	// Its records are deleted by the database too (ON DELETE CASCADE)
	delete_err := a.DeleteVehicle(id)
	if delete_err != nil {
		log.Printf("Error deleting vehicle %s: %s\n", id, delete_err)
		a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al intentar eliminar el vehículo.")
		return
	}

	if vehicle.PhotoURL != "" {
		os.Remove(strings.TrimPrefix(vehicle.PhotoURL, "/"))
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
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
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}
		fmt.Fprintf(w, "%s", buf.Bytes())
	}

	if r.Method == http.MethodPost {
		log.Println("POST method called in RecordHandler")
		result, err := a.DeleteRecord(chi.URLParam(r, "id"), chi.URLParam(r, "record_id"))
		if err != nil {
			log.Printf("Error deleting record %s of vehicle %s: %s\n", chi.URLParam(r, "record_id"), chi.URLParam(r, "id"), err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "Ocurrió un error al intentar eliminar el registro.")
			return
		}
		if result == 0 {
			a.RenderError(w, http.StatusNotFound, "Registro no encontrado.", "El registro que intentas eliminar no existe.")
			return
		}
		vehicle_page_url := fmt.Sprintf("/vehiculos/%s", chi.URLParam(r, "id"))
		http.Redirect(w, r, vehicle_page_url, http.StatusSeeOther)
	}
}

func (a App) NewRecordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		log.Println("GET called on NewRecordHandler")

		vehicle_id, conv_err := strconv.Atoi(chi.URLParam(r, "id"))
		if conv_err != nil {
			log.Println("Error with vehicle id:", conv_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "Error con ID de vehículo. ID inválido.")
			return
		}

		// For the form's warning when the km doesn't fit the vehicle's other readings
		readings, err := a.GetOdometerReadings(chi.URLParam(r, "id"), 0)
		if err != nil {
			log.Println("Error getting odometer readings:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}

		// new_record.html reads the Record fields, .Readings and .RepairCauses at the same level
		data := struct {
			Record
			Readings     []OdometerReading
			RepairCauses []string
		}{Record{VehicleID: vehicle_id}, readings, RepairCauses}

		// Display form for new record
		var buf bytes.Buffer
		err = a.tmpl.ExecuteTemplate(&buf, "new_record.html", data)
		if err != nil {
			log.Println("Error loading template:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}
		fmt.Fprintf(w, "%s", buf.Bytes())
	}

	if r.Method == http.MethodPost {
		log.Println("POST called on NewRecordHandler")

		// process data to add new record
		vehicle_id, conv_err := strconv.Atoi(chi.URLParam(r, "id"))
		if conv_err != nil {
			log.Println("Error with vehicle id:", conv_err)
			a.RenderError(w, http.StatusBadRequest, "Error del servidor.", "Error con ID de vehículo. ID inválido.")
			return
		}

		record, form_err := ParseRecordForm(r)
		if form_err != nil {
			a.RenderError(w, http.StatusBadRequest, "Datos inválidos.", form_err.Error())
			return
		}
		record.VehicleID = vehicle_id
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

func (a *App) NewTechnicianHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		log.Println("GET called on NewTechnicianHandler")

		var buf bytes.Buffer
		var technician Technician
		if err := a.tmpl.ExecuteTemplate(&buf, "new_technician.html", technician); err != nil {
			log.Println("Error rendering new_technician.html:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}
		buf.WriteTo(w)
	}

	if r.Method == http.MethodPost {
		log.Println("POST called on NewTechnicianHandler")

		technician := Technician{
			FirstName: NormalizeName(r.FormValue("first_name")),
			LastName:  NormalizeName(r.FormValue("last_name")),
		}
		if technician.FirstName == "" || technician.LastName == "" {
			a.RenderError(w, http.StatusBadRequest, "Datos inválidos.", "El nombre y el apellido son obligatorios.")
			return
		}

		insert_err := a.InsertNewTechnician(technician)
		if insert_err != nil {
			log.Println("Error inserting technician:", insert_err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Error al agregar técnico.")
			return
		}
		// back to the technicians view of the index page
		http.Redirect(w, r, "/#tecnicos", http.StatusSeeOther)
	}
}

func (a App) EditRecordHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	record_id := chi.URLParam(r, "record_id")

	record, err := a.GetRecord(id, record_id)
	if err != nil {
		log.Println("Error getting record:", err)
		a.RenderError(w, http.StatusNotFound, "Registro no encontrado.", "No se encontró el registro.")
		return
	}

	if r.Method == http.MethodGet {
		// Leaves this record out, so its own current km doesn't count as a conflict
		readings, err := a.GetOdometerReadings(id, record.ID)
		if err != nil {
			log.Println("Error getting odometer readings:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}

		// edit_record.html reads the Record fields, .Readings and .RepairCauses at the same level
		data := struct {
			Record
			Readings     []OdometerReading
			RepairCauses []string
		}{record, readings, RepairCauses}

		var buf bytes.Buffer
		err = a.tmpl.ExecuteTemplate(&buf, "edit_record.html", data)
		if err != nil {
			log.Println("Error loading template:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}
		fmt.Fprintf(w, "%s", buf.Bytes())
	}

	if r.Method == http.MethodPost {
		edited, form_err := ParseRecordForm(r)
		if form_err != nil {
			a.RenderError(w, http.StatusBadRequest, "Datos inválidos.", form_err.Error())
			return
		}
		edited.ID = record.ID
		edited.VehicleID = record.VehicleID

		update_err := a.UpdateRecord(edited)
		if update_err != nil {
			log.Println("Error updating record:", update_err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Error al editar el registro.")
			return
		}
		// back to the record's page
		http.Redirect(w, r, fmt.Sprintf("/vehiculos/%s/registro/%s", id, record_id), http.StatusSeeOther)
	}
}

func (a *App) EditTechnicianHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	technician, err := a.GetTechnician(id)
	if err != nil {
		log.Printf("Error getting technician with id: %s: %s\n", id, err)
		a.RenderError(w, http.StatusNotFound, "Técnico no encontrado.", "No se encontró el técnico.")
		return
	}

	if r.Method == http.MethodGet {
		var buf bytes.Buffer
		if err := a.tmpl.ExecuteTemplate(&buf, "edit_technician.html", technician); err != nil {
			log.Println("Error rendering edit_technician.html:", err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cargar la página.")
			return
		}
		buf.WriteTo(w)
	}

	if r.Method == http.MethodPost {
		technician.FirstName = NormalizeName(r.FormValue("first_name"))
		technician.LastName = NormalizeName(r.FormValue("last_name"))
		if technician.FirstName == "" || technician.LastName == "" {
			a.RenderError(w, http.StatusBadRequest, "Datos inválidos.", "El nombre y el apellido son obligatorios.")
			return
		}

		update_err := a.UpdateTechnician(technician)
		if update_err != nil {
			log.Println("Error updating technician:", update_err)
			a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Error al editar técnico.")
			return
		}
		http.Redirect(w, r, "/#tecnicos", http.StatusSeeOther)
	}
}

func (a *App) DeleteTechnicianHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := a.GetTechnician(id); err != nil {
		log.Printf("Error getting technician with id: %s: %s\n", id, err)
		a.RenderError(w, http.StatusNotFound, "Técnico no encontrado.", "No se encontró el técnico.")
		return
	}

	// With records attributed to them, deleting would erase who had those vehicles:
	// "dar de baja" is the way to go (see SetTechnicianActive)
	count, err := a.CountTechnicianRecords(id)
	if err != nil {
		log.Printf("Error counting records of technician %s: %s\n", id, err)
		a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al intentar eliminar el técnico.")
		return
	}
	if count > 0 {
		message := fmt.Sprintf("Tiene %d registro(s) de servicios o reparaciones a su nombre. Para no perder ese historial, dalo de baja en lugar de eliminarlo.", count)
		a.RenderError(w, http.StatusConflict, "No se puede eliminar este técnico.", message)
		return
	}

	// Their vehicle, if any, is left "Sin asignar" (see DeleteTechnician)
	if err := a.DeleteTechnician(id); err != nil {
		log.Printf("Error deleting technician %s: %s\n", id, err)
		a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al intentar eliminar el técnico.")
		return
	}
	http.Redirect(w, r, "/#tecnicos", http.StatusSeeOther)
}

// DeactivateTechnicianHandler "da de baja" a technician, and ReactivateTechnicianHandler
// undoes it. Both only change the active flag (see SetTechnicianActive).
func (a *App) DeactivateTechnicianHandler(w http.ResponseWriter, r *http.Request) {
	a.setTechnicianActive(w, r, false)
}

func (a *App) ReactivateTechnicianHandler(w http.ResponseWriter, r *http.Request) {
	a.setTechnicianActive(w, r, true)
}

func (a *App) setTechnicianActive(w http.ResponseWriter, r *http.Request, active bool) {
	id := chi.URLParam(r, "id")

	if _, err := a.GetTechnician(id); err != nil {
		log.Printf("Error getting technician with id: %s: %s\n", id, err)
		a.RenderError(w, http.StatusNotFound, "Técnico no encontrado.", "No se encontró el técnico.")
		return
	}

	if err := a.SetTechnicianActive(id, active); err != nil {
		log.Printf("Error setting technician %s active=%t: %s\n", id, active, err)
		a.RenderError(w, http.StatusInternalServerError, "Error del servidor.", "Ocurrió un error al cambiar el estado del técnico.")
		return
	}
	http.Redirect(w, r, "/#tecnicos", http.StatusSeeOther)
}
