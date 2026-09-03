package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newTestServer spins up the real router (same wiring as production, via
// NewRouter) against a fresh temp SQLite database and the real templates.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	a := newTestDB(t)

	tmpl, err := ParseTemplates()
	if err != nil {
		t.Fatalf("ParseTemplates: %v", err)
	}
	a.tmpl = tmpl

	ts := httptest.NewServer(NewRouter(a))
	t.Cleanup(ts.Close)
	return ts
}

// noRedirectClient doesn't follow redirects, so tests can inspect a
// handler's 303 + Location response directly.
func noRedirectClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func buildMultipart(t *testing.T, fields map[string]string, fileField, filename string, fileContent []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("WriteField(%s): %v", k, err)
		}
	}
	if fileField != "" {
		part, err := w.CreateFormFile(fileField, filename)
		if err != nil {
			t.Fatalf("CreateFormFile: %v", err)
		}
		if _, err := part.Write(fileContent); err != nil {
			t.Fatalf("write file content: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	return &buf, w.FormDataContentType()
}

func TestRecordHandlerRejectsWrongVehicle(t *testing.T) {
	ts := newTestServer(t)

	body, ct := buildMultipart(t, map[string]string{
		"plate": "AAA-111", "maker": "Chevrolet", "model": "Spark", "year": "2018",
	}, "", "", nil)
	resp, err := http.Post(ts.URL+"/vehiculos/new", ct, body)
	if err != nil {
		t.Fatalf("create vehicle 1: %v", err)
	}
	resp.Body.Close()

	body, ct = buildMultipart(t, map[string]string{
		"plate": "BBB-222", "maker": "Ford", "model": "Ikon", "year": "2019",
	}, "", "", nil)
	resp, err = http.Post(ts.URL+"/vehiculos/new", ct, body)
	if err != nil {
		t.Fatalf("create vehicle 2: %v", err)
	}
	resp.Body.Close()

	form := url.Values{"type": {"Servicio"}, "date": {"2026-01-01"}, "description": {"x"}}
	resp, err = http.PostForm(ts.URL+"/vehiculos/1/nuevo-registro", form)
	if err != nil {
		t.Fatalf("create record: %v", err)
	}
	resp.Body.Close()

	// Record 1 belongs to vehicle 1; asking for it under vehicle 2 must 404.
	resp, err = http.Get(ts.URL + "/vehiculos/2/registro/1")
	if err != nil {
		t.Fatalf("GET wrong-vehicle record: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}

	// Same record under its real vehicle works.
	resp2, err := http.Get(ts.URL + "/vehiculos/1/registro/1")
	if err != nil {
		t.Fatalf("GET correct-vehicle record: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp2.StatusCode, http.StatusOK)
	}
}

func TestDeleteRecordHandlerScopedToVehicle(t *testing.T) {
	ts := newTestServer(t)
	client := noRedirectClient()

	body, ct := buildMultipart(t, map[string]string{
		"plate": "AAA-111", "maker": "Chevrolet", "model": "Spark", "year": "2018",
	}, "", "", nil)
	resp, _ := http.Post(ts.URL+"/vehiculos/new", ct, body)
	resp.Body.Close()
	body, ct = buildMultipart(t, map[string]string{
		"plate": "BBB-222", "maker": "Ford", "model": "Ikon", "year": "2019",
	}, "", "", nil)
	resp, _ = http.Post(ts.URL+"/vehiculos/new", ct, body)
	resp.Body.Close()

	form := url.Values{"type": {"Servicio"}, "date": {"2026-01-01"}, "description": {"x"}}
	resp, _ = http.PostForm(ts.URL+"/vehiculos/1/nuevo-registro", form)
	resp.Body.Close()

	// Deleting record 1 through vehicle 2's URL must not remove it.
	resp, err := client.Post(ts.URL+"/vehiculos/2/registro/1/eliminar", "application/x-www-form-urlencoded", nil)
	if err != nil {
		t.Fatalf("wrong-vehicle delete: %v", err)
	}
	resp.Body.Close()

	getResp, err := http.Get(ts.URL + "/vehiculos/1/registro/1")
	if err != nil {
		t.Fatalf("GET after wrong-vehicle delete: %v", err)
	}
	getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("record should survive a wrong-vehicle delete, status = %d", getResp.StatusCode)
	}

	// Deleting through the correct vehicle removes it and redirects.
	resp, err = client.Post(ts.URL+"/vehiculos/1/registro/1/eliminar", "application/x-www-form-urlencoded", nil)
	if err != nil {
		t.Fatalf("correct-vehicle delete: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusSeeOther)
	}
	if loc := resp.Header.Get("Location"); loc != "/vehiculos/1/" {
		t.Errorf("Location = %q, want %q", loc, "/vehiculos/1/")
	}

	getResp, err = http.Get(ts.URL + "/vehiculos/1/registro/1")
	if err != nil {
		t.Fatalf("GET after correct-vehicle delete: %v", err)
	}
	getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("status after delete = %d, want %d", getResp.StatusCode, http.StatusNotFound)
	}
}

func TestNewVehicleHandlerRejectsMissingRequiredField(t *testing.T) {
	ts := newTestServer(t)

	body, ct := buildMultipart(t, map[string]string{
		"maker": "Chevrolet", "model": "Spark", "year": "2018", // no plate
	}, "", "", nil)
	resp, err := http.Post(ts.URL+"/vehiculos/new", ct, body)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestNewRecordHandlerRejectsInvalidDate(t *testing.T) {
	ts := newTestServer(t)

	body, ct := buildMultipart(t, map[string]string{
		"plate": "AAA-111", "maker": "Chevrolet", "model": "Spark", "year": "2018",
	}, "", "", nil)
	resp, _ := http.Post(ts.URL+"/vehiculos/new", ct, body)
	resp.Body.Close()

	form := url.Values{"type": {"Servicio"}, "date": {"31/02/2026"}, "description": {"x"}}
	resp, err := http.PostForm(ts.URL+"/vehiculos/1/nuevo-registro", form)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestEditVehicleHandlerPrefillsAndUpdates(t *testing.T) {
	ts := newTestServer(t) // parses templates via a relative path, so set up before chdir
	t.Chdir(t.TempDir())   // EditVehicleHandler may write photos under "uploads/"
	client := noRedirectClient()

	body, ct := buildMultipart(t, map[string]string{
		"plate": "AAA-111", "maker": "Chevrolet", "model": "Spark", "year": "2018", "location": "Morelia",
	}, "", "", nil)
	resp, _ := http.Post(ts.URL+"/vehiculos/new", ct, body)
	resp.Body.Close()

	getResp, err := http.Get(ts.URL + "/vehiculos/1/editar")
	if err != nil {
		t.Fatalf("GET edit form: %v", err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(getResp.Body)
	getResp.Body.Close()
	if !strings.Contains(buf.String(), `value="Morelia"`) {
		t.Errorf("edit form should prefill current location, got:\n%s", buf.String())
	}

	body, ct = buildMultipart(t, map[string]string{
		"plate": "AAA-111", "maker": "Chevrolet", "model": "Spark", "year": "2018", "location": "Uruapan",
	}, "photo", "car.jpg", []byte("fake jpeg bytes"))
	resp, err = client.Post(ts.URL+"/vehiculos/1/editar", ct, body)
	if err != nil {
		t.Fatalf("POST edit with photo: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusSeeOther)
	}

	getResp, err = http.Get(ts.URL + "/vehiculos/1")
	if err != nil {
		t.Fatalf("GET vehicle after edit: %v", err)
	}
	buf = new(bytes.Buffer)
	buf.ReadFrom(getResp.Body)
	getResp.Body.Close()
	page := buf.String()
	if !strings.Contains(page, "Uruapan") {
		t.Errorf("vehicle page should show updated location, got:\n%s", page)
	}
	if !strings.Contains(page, "/uploads/") {
		t.Errorf("vehicle page should show the uploaded photo, got:\n%s", page)
	}

	// Now delete the photo without uploading a new one.
	body, ct = buildMultipart(t, map[string]string{
		"plate": "AAA-111", "maker": "Chevrolet", "model": "Spark", "year": "2018",
		"location": "Uruapan", "delete_photo": "on",
	}, "", "", nil)
	resp, err = client.Post(ts.URL+"/vehiculos/1/editar", ct, body)
	if err != nil {
		t.Fatalf("POST edit with delete_photo: %v", err)
	}
	resp.Body.Close()

	getResp, err = http.Get(ts.URL + "/vehiculos/1")
	if err != nil {
		t.Fatalf("GET vehicle after photo delete: %v", err)
	}
	buf = new(bytes.Buffer)
	buf.ReadFrom(getResp.Body)
	getResp.Body.Close()
	if strings.Contains(buf.String(), "/uploads/") {
		t.Errorf("vehicle page should no longer show a photo after deletion, got:\n%s", buf.String())
	}
}
