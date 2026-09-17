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
		technicians = append(technicians, *t)
	}
	return technicians, nil
}
