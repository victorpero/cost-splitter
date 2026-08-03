package web

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestServerGetRendersUploadForm(t *testing.T) {
	server := newTestServer(t)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	if !strings.Contains(body, "Analyze CSV Files") {
		t.Fatalf("body did not contain upload form heading")
	}
	if !strings.Contains(body, "MAXI ICA") {
		t.Fatalf("body did not contain default grocery prefixes")
	}
}

func TestServerPostAnalyzesUploadedCSV(t *testing.T) {
	server := newTestServer(t)
	body, contentType := multipartRequestBody(t, map[string]string{
		"currency":       "SEK",
		"show_unmatched": "on",
		"prefixes":       "ICA\nCOOP",
		"activity.csv":   "Datum;Beskrivning;Belopp\n2026-05-11;COOP RADHUSET;-111,00\n2026-05-12;APOTEKET;50,00\n",
	})
	request := httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d\nbody:\n%s", response.Code, http.StatusOK, response.Body.String())
	}
	bodyText := response.Body.String()
	for _, want := range []string{
		"COOP RADHUSET",
		"APOTEKET",
		"-SEK 111,00",
		"-SEK 55,50",
		"SEK 50,00",
		"1 included",
		"1 unmatched",
		"Include selected",
	} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("body did not contain %q\nbody:\n%s", want, bodyText)
		}
	}
}

func TestServerPostIncludesSelectedUnmatchedTransactions(t *testing.T) {
	server := newTestServer(t)
	body, contentType := multipartRequestBody(t, map[string]string{
		"currency":       "SEK",
		"show_unmatched": "on",
		"prefixes":       "ICA\nCOOP",
		"activity.csv":   "Datum;Beskrivning;Belopp\n2026-05-11;COOP RADHUSET;-111,00\n2026-05-12;APOTEKET;50,00\n",
	})
	request := httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("initial status = %d, want %d\nbody:\n%s", response.Code, http.StatusOK, response.Body.String())
	}
	state := hiddenFieldValue(t, response.Body.String(), "transactions_state")

	body, contentType = multipartRequestBody(t, map[string]string{
		"currency":           "SEK",
		"show_unmatched":     "on",
		"prefixes":           "ICA\nCOOP",
		"transactions_state": state,
		"include_tx":         "1",
	})
	request = httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", contentType)
	response = httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("include status = %d, want %d\nbody:\n%s", response.Code, http.StatusOK, response.Body.String())
	}
	bodyText := response.Body.String()
	for _, want := range []string{
		"COOP RADHUSET",
		"APOTEKET",
		"-SEK 61,00",
		"-SEK 30,50",
		"2 included",
		"0 unmatched",
		`name="included_tx" value="1"`,
	} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("body did not contain %q\nbody:\n%s", want, bodyText)
		}
	}
}

func TestServerPostRemovesSelectedMatchedTransactions(t *testing.T) {
	server := newTestServer(t)
	body, contentType := multipartRequestBody(t, map[string]string{
		"currency":       "SEK",
		"show_unmatched": "on",
		"prefixes":       "ICA\nCOOP",
		"activity.csv":   "Datum;Beskrivning;Belopp\n2026-05-11;COOP RADHUSET;-111,00\n2026-05-12;APOTEKET;50,00\n",
	})
	request := httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("initial status = %d, want %d\nbody:\n%s", response.Code, http.StatusOK, response.Body.String())
	}
	state := hiddenFieldValue(t, response.Body.String(), "transactions_state")

	body, contentType = multipartRequestBody(t, map[string]string{
		"currency":           "SEK",
		"show_unmatched":     "on",
		"prefixes":           "ICA\nCOOP",
		"transactions_state": state,
		"remove_tx":          "0",
	})
	request = httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", contentType)
	response = httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("remove status = %d, want %d\nbody:\n%s", response.Code, http.StatusOK, response.Body.String())
	}
	bodyText := response.Body.String()
	for _, want := range []string{
		"COOP RADHUSET",
		"APOTEKET",
		"SEK 0,00",
		"0 included",
		"2 unmatched",
		`name="excluded_tx" value="0"`,
	} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("body did not contain %q\nbody:\n%s", want, bodyText)
		}
	}
}

func TestServerPostStoresAllocationsIndependently(t *testing.T) {
	server := newTestServer(t)
	body, contentType := multipartRequestBody(t, map[string]string{
		"currency":     "SEK",
		"prefixes":     "ICA\nCOOP",
		"activity.csv": "Datum;Beskrivning;Belopp\n2026-05-11;ICA ONE;-100,00\n2026-05-12;COOP TWO;40,00\n",
	})
	request := httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("initial status = %d, want %d\nbody:\n%s", response.Code, http.StatusOK, response.Body.String())
	}
	state := hiddenFieldValue(t, response.Body.String(), "transactions_state")

	body, contentType = multipartRequestBody(t, map[string]string{
		"currency":           "SEK",
		"prefixes":           "ICA\nCOOP",
		"transactions_state": state,
		"allocation_tx_0":    "participant_one",
	})
	request = httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", contentType)
	response = httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("first allocation status = %d, want %d\nbody:\n%s", response.Code, http.StatusOK, response.Body.String())
	}
	bodyText := response.Body.String()
	if got := selectedAllocation(t, bodyText, "0"); got != "participant_one" {
		t.Fatalf("allocation for transaction 0 = %q, want participant_one", got)
	}
	if got := selectedAllocation(t, bodyText, "1"); got != "split_evenly" {
		t.Fatalf("allocation for transaction 1 = %q, want split_evenly", got)
	}
	for _, want := range []string{"-SEK 60,00", "-SEK 80,00", "SEK 20,00"} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("body did not contain %q\nbody:\n%s", want, bodyText)
		}
	}
	state = hiddenFieldValue(t, bodyText, "transactions_state")

	body, contentType = multipartRequestBody(t, map[string]string{
		"currency":           "SEK",
		"prefixes":           "ICA\nCOOP",
		"transactions_state": state,
		"allocation_tx_1":    "participant_two",
	})
	request = httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", contentType)
	response = httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("second allocation status = %d, want %d\nbody:\n%s", response.Code, http.StatusOK, response.Body.String())
	}
	bodyText = response.Body.String()
	if got := selectedAllocation(t, bodyText, "0"); got != "participant_one" {
		t.Fatalf("allocation for transaction 0 after transaction 1 changed = %q, want participant_one", got)
	}
	if got := selectedAllocation(t, bodyText, "1"); got != "participant_two" {
		t.Fatalf("allocation for transaction 1 = %q, want participant_two", got)
	}
	for _, want := range []string{"-SEK 60,00", "-SEK 100,00", "SEK 40,00"} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("body did not contain %q\nbody:\n%s", want, bodyText)
		}
	}
}

func TestServerPostStoresAmountsToSplitIndependently(t *testing.T) {
	server := newTestServer(t)
	body, contentType := multipartRequestBody(t, map[string]string{
		"currency":     "SEK",
		"prefixes":     "ICA\nCOOP",
		"activity.csv": "Datum;Beskrivning;Belopp\n2026-05-11;ICA ONE;-1500,00\n2026-05-12;COOP TWO;-500,00\n",
	})
	request := httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("initial status = %d, want %d\nbody:\n%s", response.Code, http.StatusOK, response.Body.String())
	}
	state := hiddenFieldValue(t, response.Body.String(), "transactions_state")

	body, contentType = multipartRequestBody(t, map[string]string{
		"currency":           "SEK",
		"prefixes":           "ICA\nCOOP",
		"transactions_state": state,
		"split_amount_tx_0":  "-600,00",
	})
	request = httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", contentType)
	response = httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("first adjustment status = %d, want %d\nbody:\n%s", response.Code, http.StatusOK, response.Body.String())
	}
	bodyText := response.Body.String()
	if got := splitAmountValue(t, bodyText, "0"); got != "-600,00" {
		t.Fatalf("amount to split for transaction 0 = %q, want -600,00", got)
	}
	if got := splitAmountValue(t, bodyText, "1"); got != "-500,00" {
		t.Fatalf("amount to split for transaction 1 = %q, want -500,00", got)
	}
	for _, want := range []string{"Imported Amount", "Amount to Split", "-SEK 1 500,00", "-SEK 1 100,00", "-SEK 550,00"} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("body did not contain %q\nbody:\n%s", want, bodyText)
		}
	}
	state = hiddenFieldValue(t, bodyText, "transactions_state")

	body, contentType = multipartRequestBody(t, map[string]string{
		"currency":           "SEK",
		"prefixes":           "ICA\nCOOP",
		"transactions_state": state,
		"split_amount_tx_1":  "-200,00",
	})
	request = httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", contentType)
	response = httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("second adjustment status = %d, want %d\nbody:\n%s", response.Code, http.StatusOK, response.Body.String())
	}
	bodyText = response.Body.String()
	if got := splitAmountValue(t, bodyText, "0"); got != "-600,00" {
		t.Fatalf("amount to split for transaction 0 after transaction 1 changed = %q, want -600,00", got)
	}
	if got := splitAmountValue(t, bodyText, "1"); got != "-200,00" {
		t.Fatalf("amount to split for transaction 1 = %q, want -200,00", got)
	}
	for _, want := range []string{"-SEK 800,00", "-SEK 400,00"} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("body did not contain %q\nbody:\n%s", want, bodyText)
		}
	}
}

func TestServerDoesNotRenderRemovedAmountMode(t *testing.T) {
	server := newTestServer(t)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	body := response.Body.String()
	for _, unwanted := range []string{"amount_mode", "Signed CSV amounts"} {
		if strings.Contains(body, unwanted) {
			t.Fatalf("body unexpectedly contained removed amount-mode UI %q", unwanted)
		}
	}
}

func TestServerPostRequiresCSVFile(t *testing.T) {
	server := newTestServer(t)
	body, contentType := multipartRequestBody(t, map[string]string{
		"currency": "SEK",
		"prefixes": "ICA\nCOOP",
	})
	request := httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), "Choose at least one American Express CSV file") {
		t.Fatalf("body did not contain missing file error")
	}
}

func hiddenFieldValue(t *testing.T, body string, name string) string {
	t.Helper()
	pattern := regexp.MustCompile(`name="` + regexp.QuoteMeta(name) + `" value="([^"]+)"`)
	matches := pattern.FindStringSubmatch(body)
	if len(matches) != 2 {
		t.Fatalf("body did not contain hidden field %q\nbody:\n%s", name, body)
	}
	return matches[1]
}

func selectedAllocation(t *testing.T, body string, id string) string {
	t.Helper()
	pattern := regexp.MustCompile(`(?s)<select name="allocation_tx_` + regexp.QuoteMeta(id) + `"[^>]*>(.*?)</select>`)
	selectMatch := pattern.FindStringSubmatch(body)
	if len(selectMatch) != 2 {
		t.Fatalf("body did not contain allocation select for %q\nbody:\n%s", id, body)
	}
	selected := regexp.MustCompile(`<option value="([^"]+)" selected`).FindStringSubmatch(selectMatch[1])
	if len(selected) != 2 {
		t.Fatalf("allocation select for %q did not contain a selected option\nselect:\n%s", id, selectMatch[1])
	}
	return selected[1]
}

func splitAmountValue(t *testing.T, body string, id string) string {
	t.Helper()
	pattern := regexp.MustCompile(`name="split_amount_tx_` + regexp.QuoteMeta(id) + `" value="([^"]+)"`)
	matches := pattern.FindStringSubmatch(body)
	if len(matches) != 2 {
		t.Fatalf("body did not contain amount-to-split input for %q\nbody:\n%s", id, body)
	}
	return matches[1]
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	server, err := NewServer(Config{Currency: "SEK"})
	if err != nil {
		t.Fatalf("NewServer returned error: %v", err)
	}
	return server
}

func multipartRequestBody(t *testing.T, values map[string]string) (io.Reader, string) {
	t.Helper()
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	for name, value := range values {
		if strings.HasSuffix(name, ".csv") {
			part, err := writer.CreateFormFile("files", name)
			if err != nil {
				t.Fatalf("CreateFormFile returned error: %v", err)
			}
			if _, err := part.Write([]byte(value)); err != nil {
				t.Fatalf("write file part returned error: %v", err)
			}
			continue
		}
		if err := writer.WriteField(name, value); err != nil {
			t.Fatalf("WriteField returned error: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart writer close returned error: %v", err)
	}

	return &buffer, writer.FormDataContentType()
}
