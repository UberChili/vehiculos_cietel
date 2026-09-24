package main

type Technician struct {
	ID        int
	FirstName string
	LastName  string
}

func (a *App) findAllTechnicians() ([]Technician, error) {
	query := `SELECT id, first_name, last_name FROM technicians ORDER BY last_name`

	rows, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var technicians []Technician

	for rows.Next() {
		t := &Technician{}
		err := rows.Scan(&t.ID, &t.FirstName, &t.LastName)

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
	query := `SELECT id, first_name, last_name FROM technicians WHERE id = ?`

	t := Technician{}
	err := a.db.QueryRow(query, id).Scan(&t.ID, &t.FirstName, &t.LastName)
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
