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

// InsertNewTechnician adds a technician with no vehicle. Assigning one is done
// from the vehicle's side (vehicles.assigned_to points to technicians.id).
func (a *App) InsertNewTechnician(technician Technician) error {
	stmt := `INSERT INTO technicians (first_name, last_name) VALUES (?, ?)`

	_, err := a.db.Exec(stmt, technician.FirstName, technician.LastName)
	return err
}
