package main

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

// formRequest builds a request with both PostForm and Form pre-populated,
// so PostFormValue/FormValue read straight from it without trying (and
// possibly failing) to parse a real body.
func formRequest(values map[string]string) *http.Request {
	form := url.Values{}
	for k, v := range values {
		form.Set(k, v)
	}
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.PostForm = form
	req.Form = form
	return req
}

func wantValidationError(t *testing.T, err error, label string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected a ValidationError, got nil", label)
	}
	var verr *ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("%s: expected a ValidationError, got %T: %v", label, err, err)
	}
}

func TestRecordValidate(t *testing.T) {
	base := Record{DateShort: "2026-09-02", Type: "Servicio", Description: "Cambio de aceite"}

	t.Run("valid record passes", func(t *testing.T) {
		if err := base.Validate(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("december is a valid month", func(t *testing.T) {
		// Regression: an earlier hand-rolled validator rejected month 12
		// with an off-by-one (`month >= 12` instead of `> 12`).
		r := base
		r.DateShort = "2026-12-31"
		if err := r.Validate(); err != nil {
			t.Fatalf("expected December 31st to be valid, got %v", err)
		}
	})

	t.Run("empty date", func(t *testing.T) {
		r := base
		r.DateShort = ""
		wantValidationError(t, r.Validate(), "empty date")
	})

	t.Run("slash-separated date is rejected, not just miscounted", func(t *testing.T) {
		// Regression: the original implementation split on "/" assuming
		// DD/MM/YYYY, but <input type="date"> always submits ISO
		// (YYYY-MM-DD), so a slash-separated value should fail cleanly
		// instead of panicking on an out-of-range index.
		r := base
		r.DateShort = "02/09/2026"
		wantValidationError(t, r.Validate(), "slash-separated date")
	})

	t.Run("calendar-invalid date", func(t *testing.T) {
		r := base
		r.DateShort = "2026-02-30" // February never has a 30th
		wantValidationError(t, r.Validate(), "Feb 30")
	})

	t.Run("missing type", func(t *testing.T) {
		r := base
		r.Type = ""
		wantValidationError(t, r.Validate(), "missing type")
	})

	t.Run("missing description", func(t *testing.T) {
		r := base
		r.Description = ""
		wantValidationError(t, r.Validate(), "missing description")
	})
}

func TestVehicleValidate(t *testing.T) {
	base := Vehicle{Plate: "ABC-123", Maker: "Chevrolet", Model: "Spark", Year: "2018"}

	t.Run("valid vehicle passes", func(t *testing.T) {
		if err := base.Validate(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	for _, field := range []string{"Plate", "Maker", "Model", "Year"} {
		t.Run("missing "+field, func(t *testing.T) {
			v := base
			switch field {
			case "Plate":
				v.Plate = "  "
			case "Maker":
				v.Maker = "  "
			case "Model":
				v.Model = "  "
			case "Year":
				v.Year = "  "
			}
			wantValidationError(t, v.Validate(), "missing "+field)
		})
	}
}

func TestNewVehicleFromFormSanitizes(t *testing.T) {
	req := formRequest(map[string]string{
		"plate":       "abc-123",
		"maker":       "chevrolet",
		"model":       "spark",
		"year":        "2018",
		"assigned_to": "juan pérez",
		"location":    "morelia",
	})

	v := NewVehicleFromForm(req)

	if v.Plate != "ABC-123" {
		t.Errorf("Plate = %q, want uppercased %q", v.Plate, "ABC-123")
	}
	if v.Maker != "Chevrolet" {
		t.Errorf("Maker = %q, want %q", v.Maker, "Chevrolet")
	}
	if v.Model != "Spark" {
		t.Errorf("Model = %q, want %q", v.Model, "Spark")
	}
	if v.AssignedTo != "Juan pérez" {
		t.Errorf("AssignedTo = %q, want %q", v.AssignedTo, "Juan pérez")
	}
	if v.Location != "Morelia" {
		t.Errorf("Location = %q, want %q", v.Location, "Morelia")
	}
}

func TestNewVehicleFromFormLeavesOptionalFieldsEmpty(t *testing.T) {
	req := formRequest(map[string]string{
		"plate": "abc-123",
		"maker": "chevrolet",
		"model": "spark",
		"year":  "2018",
	})

	v := NewVehicleFromForm(req)

	if v.AssignedTo != "" {
		t.Errorf("AssignedTo = %q, want empty", v.AssignedTo)
	}
	if v.Location != "" {
		t.Errorf("Location = %q, want empty", v.Location)
	}
}

func TestNewRecordFromForm(t *testing.T) {
	// The date field (new_record.html) submits dd-mm-yyyy, matching how
	// dates are displayed everywhere else in the app; NewRecordFromForm
	// must convert it to ISO before it reaches storage/validation.
	req := formRequest(map[string]string{
		"date":        "02-09-2026",
		"type":        "Reparación",
		"description": "Cambio de balatas",
		"cost":        "1500.00",
	})

	r := NewRecordFromForm(req)

	if r.DateShort != "2026-09-02" || r.Type != "Reparación" ||
		r.Description != "Cambio de balatas" || r.Cost != "1500.00" {
		t.Errorf("unexpected record from form: %+v", r)
	}
}

func TestNewRecordFromFormLeavesUnparseableDateForValidateToReject(t *testing.T) {
	req := formRequest(map[string]string{
		"date":        "not-a-date",
		"type":        "Servicio",
		"description": "x",
	})

	r := NewRecordFromForm(req)

	if r.DateShort != "not-a-date" {
		t.Errorf("DateShort = %q, want unchanged %q", r.DateShort, "not-a-date")
	}
	if err := r.Validate(); err == nil {
		t.Error("expected Validate to reject the unparseable date")
	}
}

// multipartPhotoRequest builds a real multipart/form-data request with a
// "photo" file field, so SavePhoto's r.FormFile call has something to parse.
func multipartPhotoRequest(t *testing.T, filename string, content []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if filename != "" {
		part, err := w.CreateFormFile("photo", filename)
		if err != nil {
			t.Fatalf("CreateFormFile: %v", err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatalf("write photo content: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if err := req.ParseMultipartForm(10 << 20); err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}
	return req
}

func TestSavePhotoNoFile(t *testing.T) {
	req := multipartPhotoRequest(t, "", nil)
	photoURL, err := SavePhoto(req)
	if err != nil {
		t.Fatalf("expected no error when no photo is submitted, got %v", err)
	}
	if photoURL != "" {
		t.Errorf("expected empty URL when no photo is submitted, got %q", photoURL)
	}
}

func TestSavePhotoRejectsDisallowedExtension(t *testing.T) {
	req := multipartPhotoRequest(t, "malware.exe", []byte("not really an exe"))
	_, err := SavePhoto(req)
	wantValidationError(t, err, "disallowed extension")
}

func TestSavePhotoAcceptsAllowedExtension(t *testing.T) {
	t.Chdir(t.TempDir()) // SavePhoto writes under a relative "uploads/" dir

	req := multipartPhotoRequest(t, "car.jpg", []byte("fake jpeg bytes"))
	got, err := SavePhoto(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got == "" {
		t.Fatal("expected a non-empty URL")
	}

	diskPath := filepath.Join(uploadDir, filepath.Base(got))
	if _, err := os.Stat(diskPath); err != nil {
		t.Errorf("expected file at %s, got error: %v", diskPath, err)
	}
}

func TestDeleteUploadedPhotoRemovesFile(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(uploadDir, "delete-me-test.jpg")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	deleteUploadedPhoto("/" + uploadDir + "/delete-me-test.jpg")

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file to be removed, stat error: %v", err)
	}
}

func TestDeleteUploadedPhotoIgnoresPathsOutsideUploadDir(t *testing.T) {
	// Should be a no-op (not attempt removal) for anything that isn't
	// under the expected "/uploads/" prefix. This just needs to not
	// panic or error for suspicious/foreign input.
	deleteUploadedPhoto("../../etc/passwd")
	deleteUploadedPhoto("/etc/passwd")
	deleteUploadedPhoto("")
}
