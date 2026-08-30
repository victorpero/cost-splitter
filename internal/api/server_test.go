package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/victorpero/cost-splitter/internal/application"
)

func TestServerExposesHealthReadinessAndDefaults(t *testing.T) {
	server := newTestServer()
	for _, test := range []struct {
		path string
		want string
	}{
		{path: "/healthz", want: `"status":"ok"`},
		{path: "/readyz", want: `"status":"ready"`},
		{path: "/api/v1/defaults", want: `"currency":"SEK"`},
	} {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), test.want) {
			t.Fatalf("GET %s = %d %s", test.path, response.Code, response.Body.String())
		}
	}
}

func TestServerImportsCSVAndCalculatesSplit(t *testing.T) {
	server := newTestServer()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	file, err := writer.CreateFormFile("files", "activity.csv")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = file.Write([]byte("Datum;Beskrivning;Belopp\n2026-05-11;COOP RADHUSET;-111,00\n2026-05-12;APOTEKET;50,00\n"))
	_ = writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/api/v1/imports/amex", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("import status = %d, body = %s", response.Code, response.Body.String())
	}

	var imported struct {
		Data importResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &imported); err != nil {
		t.Fatal(err)
	}
	if imported.Data.FilesCount != 1 || imported.Data.RowsCount != 2 {
		t.Fatalf("import response = %+v", imported.Data)
	}

	calculation := calculationRequest{
		Prefixes: []string{"ICA", "COOP"},
		Transactions: []calculationTransaction{
			{transactionModel: imported.Data.Transactions[0]},
			{transactionModel: imported.Data.Transactions[1], Selection: application.SelectionIncluded},
		},
	}
	calculationBody, _ := json.Marshal(calculation)
	response = httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/split-calculations", bytes.NewReader(calculationBody)))
	if response.Code != http.StatusOK {
		t.Fatalf("calculation status = %d, body = %s", response.Code, response.Body.String())
	}
	var calculated struct {
		Data calculationResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &calculated); err != nil {
		t.Fatal(err)
	}
	if len(calculated.Data.Included) != 2 || calculated.Data.Totals.TotalCents != -6100 || calculated.Data.Totals.ParticipantOneHalfCents != -6100 {
		t.Fatalf("calculation response = %+v", calculated.Data)
	}
}

func TestServerReturnsStructuredErrors(t *testing.T) {
	server := newTestServer()
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/split-calculations", strings.NewReader(`{"unexpected":true}`)))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
	if got := response.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("Content-Type = %q", got)
	}
	for _, want := range []string{`"error"`, `"code":"invalid_json"`, `"message"`} {
		if !strings.Contains(response.Body.String(), want) {
			t.Fatalf("body = %s, want %s", response.Body.String(), want)
		}
	}
}

func TestServerAllowsConfiguredOrigins(t *testing.T) {
	server := newTestServer()
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/defaults", nil)
	request.Header.Set("Origin", "https://client.example")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "https://client.example" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}

func newTestServer() *Server {
	return NewServer(Config{
		Currency:       "SEK",
		AllowedOrigins: []string{"https://client.example"},
	}, application.NewService())
}
