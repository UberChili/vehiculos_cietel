package main

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

// newTestDB opens a fresh SQLite database in a temp directory, with the
// same schema/migrations CreateOrOpenTable applies in production.
func newTestDB(t *testing.T) *App {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := CreateOrOpenTable(path)
	if err != nil {
		t.Fatalf("CreateOrOpenTable: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &App{db: db}
}

func mustInsertVehicle(t *testing.T, a *App, v Vehicle) int {
	t.Helper()
	if err := a.InsertVehicle(v); err != nil {
		t.Fatalf("InsertVehicle: %v", err)
	}
	var id int
	if err := a.db.QueryRow("SELECT id FROM vehicles WHERE plate = ?", v.Plate).Scan(&id); err != nil {
		t.Fatalf("looking up inserted vehicle id: %v", err)
	}
	return id
}

func TestInsertAndGetVehicle(t *testing.T) {
	a := newTestDB(t)
	v := Vehicle{Plate: "ABC-123", Maker: "Chevrolet", Model: "Spark", Year: "2018", AssignedTo: "Juan", Location: "Morelia"}
	id := mustInsertVehicle(t, a, v)

	got, err := a.GetVehicleByID(id)
	if err != nil {
		t.Fatalf("GetVehicleByID: %v", err)
	}
	if got.Plate != v.Plate || got.Maker != v.Maker || got.Model != v.Model ||
		got.Year != v.Year || got.AssignedTo != v.AssignedTo || got.Location != v.Location {
		t.Errorf("got %+v, want fields matching %+v", got, v)
	}
	if got.LastService != "" || got.LastRepair != "" {
		t.Errorf("expected empty LastService/LastRepair with no records, got %q/%q", got.LastService, got.LastRepair)
	}
}

func TestInsertVehicleRejectsInvalidData(t *testing.T) {
	a := newTestDB(t)
	err := a.InsertVehicle(Vehicle{Maker: "Chevrolet", Model: "Spark", Year: "2018"}) // no plate
	var verr *ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("expected a ValidationError for missing plate, got %v", err)
	}
}

func TestGetVehicleByIDNotFound(t *testing.T) {
	a := newTestDB(t)
	_, err := a.GetVehicleByID(999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestLastServiceAndLastRepairAreDerivedFromRecords(t *testing.T) {
	a := newTestDB(t)
	id := mustInsertVehicle(t, a, Vehicle{Plate: "ABC-123", Maker: "Chevrolet", Model: "Spark", Year: "2018"})
	idStr := "1" // first vehicle inserted in a fresh db

	records := []Record{
		{DateShort: "2025-01-10", Type: "Servicio", Description: "Cambio de aceite"},
		{DateShort: "2026-06-20", Type: "Servicio", Description: "Cambio de aceite y filtros"}, // latest Servicio
		{DateShort: "2026-01-05", Type: "Reparación", Description: "Cambio de balatas"},
		{DateShort: "2025-11-30", Type: "Reparación", Description: "Cambio de amortiguadores"},
	}
	for _, r := range records {
		if err := a.InsertRecord(idStr, r); err != nil {
			t.Fatalf("InsertRecord: %v", err)
		}
	}

	got, err := a.GetVehicleByID(id)
	if err != nil {
		t.Fatalf("GetVehicleByID: %v", err)
	}
	if got.LastService != "2026-06-20" {
		t.Errorf("LastService = %q, want %q", got.LastService, "2026-06-20")
	}
	if got.LastRepair != "2026-01-05" {
		t.Errorf("LastRepair = %q, want %q", got.LastRepair, "2026-01-05")
	}
}

func TestGetVehicleRecordsByIDSortsLatestFirst(t *testing.T) {
	a := newTestDB(t)
	mustInsertVehicle(t, a, Vehicle{Plate: "ABC-123", Maker: "Chevrolet", Model: "Spark", Year: "2018"})

	dates := []string{"2025-01-10", "2026-06-20", "2025-11-30"}
	for _, d := range dates {
		if err := a.InsertRecord("1", Record{DateShort: d, Type: "Servicio", Description: "x"}); err != nil {
			t.Fatalf("InsertRecord: %v", err)
		}
	}

	got, err := a.GetVehicleRecordsByID(1)
	if err != nil {
		t.Fatalf("GetVehicleRecordsByID: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 records, got %d", len(got))
	}
	want := []string{"2026-06-20", "2025-11-30", "2025-01-10"}
	for i, w := range want {
		if got[i].DateShort != w {
			t.Errorf("position %d: got date %q, want %q", i, got[i].DateShort, w)
		}
	}
}

func TestGetRecordByIDScopedToVehicle(t *testing.T) {
	a := newTestDB(t)
	mustInsertVehicle(t, a, Vehicle{Plate: "AAA-111", Maker: "Chevrolet", Model: "Spark", Year: "2018"})
	mustInsertVehicle(t, a, Vehicle{Plate: "BBB-222", Maker: "Ford", Model: "Ikon", Year: "2019"})
	if err := a.InsertRecord("1", Record{DateShort: "2026-01-01", Type: "Servicio", Description: "x"}); err != nil {
		t.Fatalf("InsertRecord: %v", err)
	}

	if _, err := a.GetRecordByID(1, 1); err != nil {
		t.Errorf("expected to find record 1 under vehicle 1, got %v", err)
	}
	if _, err := a.GetRecordByID(2, 1); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows for record 1 under the wrong vehicle, got %v", err)
	}
}

func TestDeleteRecordScopedToVehicle(t *testing.T) {
	a := newTestDB(t)
	mustInsertVehicle(t, a, Vehicle{Plate: "AAA-111", Maker: "Chevrolet", Model: "Spark", Year: "2018"})
	mustInsertVehicle(t, a, Vehicle{Plate: "BBB-222", Maker: "Ford", Model: "Ikon", Year: "2019"})
	if err := a.InsertRecord("1", Record{DateShort: "2026-01-01", Type: "Servicio", Description: "x"}); err != nil {
		t.Fatalf("InsertRecord: %v", err)
	}

	// Deleting through the wrong vehicle must not touch the record.
	if err := a.DeleteRecord(2, 1); err != nil {
		t.Fatalf("DeleteRecord(wrong vehicle): %v", err)
	}
	if _, err := a.GetRecordByID(1, 1); err != nil {
		t.Fatalf("record should still exist after a wrong-vehicle delete attempt, got %v", err)
	}

	// Deleting through the right vehicle removes it.
	if err := a.DeleteRecord(1, 1); err != nil {
		t.Fatalf("DeleteRecord(right vehicle): %v", err)
	}
	if _, err := a.GetRecordByID(1, 1); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected record to be gone, got %v", err)
	}
}

func TestUpdateVehicle(t *testing.T) {
	a := newTestDB(t)
	id := mustInsertVehicle(t, a, Vehicle{Plate: "ABC-123", Maker: "Chevrolet", Model: "Spark", Year: "2018", Location: "Morelia"})

	updated := Vehicle{ID: id, Plate: "ABC-123", Maker: "Chevrolet", Model: "Spark", Year: "2018", Location: "Uruapan", PhotoURL: "/uploads/x.jpg"}
	if err := a.UpdateVehicle(updated); err != nil {
		t.Fatalf("UpdateVehicle: %v", err)
	}

	got, err := a.GetVehicleByID(id)
	if err != nil {
		t.Fatalf("GetVehicleByID: %v", err)
	}
	if got.Location != "Uruapan" {
		t.Errorf("Location = %q, want %q", got.Location, "Uruapan")
	}
	if got.PhotoURL != "/uploads/x.jpg" {
		t.Errorf("PhotoURL = %q, want %q", got.PhotoURL, "/uploads/x.jpg")
	}
}

func TestUpdateVehicleRejectsInvalidData(t *testing.T) {
	a := newTestDB(t)
	id := mustInsertVehicle(t, a, Vehicle{Plate: "ABC-123", Maker: "Chevrolet", Model: "Spark", Year: "2018"})

	err := a.UpdateVehicle(Vehicle{ID: id, Maker: "Chevrolet", Model: "Spark", Year: "2018"}) // no plate
	var verr *ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("expected a ValidationError for missing plate, got %v", err)
	}
}

func TestInsertRecordRejectsInvalidData(t *testing.T) {
	a := newTestDB(t)
	mustInsertVehicle(t, a, Vehicle{Plate: "ABC-123", Maker: "Chevrolet", Model: "Spark", Year: "2018"})

	err := a.InsertRecord("1", Record{DateShort: "not-a-date", Type: "Servicio", Description: "x"})
	var verr *ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("expected a ValidationError for an invalid date, got %v", err)
	}
}
