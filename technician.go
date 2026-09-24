package main

import (
	"errors"
	"fmt"
)

type Technician struct {
	ID        int
	FirstName string
	LastName  string
	Active    bool // false once "dado de baja": kept for their records' history, but can't get a vehicle
}

func (a *App) findAllTechnicians() ([]Technician, error) {
	// Active ones first, then the ones "dados de baja"
	query := `SELECT id, first_name, last_name, active FROM technicians ORDER BY active DESC, last_name`

	rows, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var technicians []Technician

	for rows.Next() {
		t := &Technician{}
		err := rows.Scan(&t.ID, &t.FirstName, &t.LastName, &t.Active)

		if err != nil {
			return nil, err
		}
		// Stored lowercase, capitalized for display
		t.FirstName = CapitalizeWords(t.FirstName)
		t.LastName = CapitalizeWords(t.LastName)
		technicians = append(technicians, *t)
	}
	return technicians, nil
}

func (a *App) GetTechnician(id string) (Technician, error) {
	query := `SELECT id, first_name, last_name, active FROM technicians WHERE id = ?`

	t := Technician{}
	err := a.db.QueryRow(query, id).Scan(&t.ID, &t.FirstName, &t.LastName, &t.Active)
	if err != nil {
		return Technician{}, err
	}
	// Stored lowercase, capitalized for display
	t.FirstName = CapitalizeWords(t.FirstName)
	t.LastName = CapitalizeWords(t.LastName)
	return t, nil
}

// InsertNewTechnician adds a technician with no vehicle. Assigning one is done
// from the vehicle's side (vehicles.assigned_to points to technicians.id).
func (a *App) InsertNewTechnician(technician Technician) error {
	stmt := `INSERT INTO technicians (first_name, last_name) VALUES (?, ?)`

	_, err := a.db.Exec(stmt, technician.FirstName, technician.LastName)
	return err
}

func (a *App) UpdateTechnician(technician Technician) error {
	stmt := `UPDATE technicians SET first_name = ?, last_name = ? WHERE id = ?`

	_, err := a.db.Exec(stmt, technician.FirstName, technician.LastName, technician.ID)
	return err
}

// DeleteTechnician removes a technician. Their vehicle, if they have one, is left
// "Sin asignar" first, since vehicles.assigned_to can't point to a deleted
// technician. Both run in one transaction, so either both happen or neither does.
func (a *App) DeleteTechnician(id string) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // does nothing if Commit already succeeded

	if _, err := tx.Exec(`UPDATE vehicles SET assigned_to = NULL WHERE assigned_to = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM technicians WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

// SetTechnicianActive "da de baja" (active = false) or reactivates a technician.
// Their records keep pointing to them, which is the point: a technician who left
// is still in the history. Going inactive also leaves their vehicle "Sin asignar",
// in the same transaction as DeleteTechnician does.
func (a *App) SetTechnicianActive(id string, active bool) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // does nothing if Commit already succeeded

	if !active {
		if _, err := tx.Exec(`UPDATE vehicles SET assigned_to = NULL WHERE assigned_to = ?`, id); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`UPDATE technicians SET active = ? WHERE id = ?`, active, id); err != nil {
		return err
	}
	return tx.Commit()
}

// CountTechnicianRecords tells how many records are attributed to a technician.
// A technician with records can only be "dado de baja", not deleted, or those
// records would lose who had the vehicle.
func (a *App) CountTechnicianRecords(id string) (int, error) {
	var count int
	err := a.db.QueryRow(`SELECT COUNT(*) FROM records WHERE technician_id = ?`, id).Scan(&count)
	return count, err
}

// CheckTechnicianAssignable makes sure a vehicle is only assigned to an existing,
// active technician. The forms only list active ones; this also covers a form
// left open while someone "dio de baja" that technician.
func (a *App) CheckTechnicianAssignable(technician_id *int) error {
	if technician_id == nil { // "Sin asignar"
		return nil
	}
	t, err := a.GetTechnician(fmt.Sprint(*technician_id))
	if err != nil {
		return errors.New("El técnico seleccionado no existe.")
	}
	if !t.Active {
		return errors.New("El técnico seleccionado está dado de baja.")
	}
	return nil
}
