package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/victorpero/cost-splitter/internal/application"
	"github.com/victorpero/cost-splitter/internal/matcher"
	"github.com/victorpero/cost-splitter/internal/split"
	"github.com/victorpero/cost-splitter/internal/transaction"
)

const (
	maxUploadBytes = 32 << 20
	maxJSONBytes   = 4 << 20
)

type Config struct {
	Currency       string
	AllowedOrigins []string
}

type Server struct {
	config  Config
	service *application.Service
	handler http.Handler
}

type envelope struct {
	Data  any       `json:"data,omitempty"`
	Error *apiError `json:"error,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type transactionModel struct {
	ID          string `json:"id"`
	Date        string `json:"date"`
	Description string `json:"description"`
	AmountCents int64  `json:"amount_cents"`
	SourceFile  string `json:"source_file,omitempty"`
	SourceLine  int    `json:"source_line,omitempty"`
}

type defaultsResponse struct {
	Currency string   `json:"currency"`
	Prefixes []string `json:"prefixes"`
}

type importResponse struct {
	FilesCount   int                `json:"files_count"`
	RowsCount    int                `json:"rows_count"`
	Transactions []transactionModel `json:"transactions"`
}

type calculationRequest struct {
	Prefixes     []string                 `json:"prefixes"`
	Transactions []calculationTransaction `json:"transactions"`
}

type calculationTransaction struct {
	transactionModel
	Selection        application.Selection `json:"selection,omitempty"`
	SplitAmountCents *int64                `json:"split_amount_cents,omitempty"`
	Allocation       split.Allocation      `json:"allocation,omitempty"`
}

type calculationResponse struct {
	Included  []calculationRow `json:"included"`
	Unmatched []calculationRow `json:"unmatched"`
	Totals    totalsModel      `json:"totals"`
}

type calculationRow struct {
	transactionModel
	SplitAmountCents int64            `json:"split_amount_cents"`
	Allocation       split.Allocation `json:"allocation"`
}

type totalsModel struct {
	TotalCents              int64 `json:"total_cents"`
	ParticipantOneHalfCents int64 `json:"participant_one_half_cents"`
	ParticipantTwoHalfCents int64 `json:"participant_two_half_cents"`
}

func NewServer(config Config, service *application.Service) *Server {
	if strings.TrimSpace(config.Currency) == "" {
		config.Currency = "SEK"
	}
	if service == nil {
		service = application.NewService()
	}

	server := &Server{config: config, service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", server.health)
	mux.HandleFunc("/readyz", server.ready)
	mux.HandleFunc("/api/v1/defaults", server.defaults)
	mux.HandleFunc("/api/v1/imports/amex", server.importAmex)
	mux.HandleFunc("/api/v1/split-calculations", server.calculateSplit)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "endpoint not found")
	})
	server.handler = server.withCORS(mux)
	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, envelope{Data: map[string]string{"status": "ok"}})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, envelope{Data: map[string]string{"status": "ready"}})
}

func (s *Server) defaults(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, envelope{Data: defaultsResponse{
		Currency: s.config.Currency,
		Prefixes: matcher.DefaultStorePrefixes(),
	}})
}

func (s *Server) importAmex(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_upload", fmt.Sprintf("could not read uploaded files: %v", err))
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	fileHeaders := r.MultipartForm.File["files"]
	if len(fileHeaders) == 0 {
		writeError(w, http.StatusBadRequest, "missing_files", "at least one CSV file is required")
		return
	}

	files, closeFiles, err := openImportFiles(fileHeaders)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_upload", err.Error())
		return
	}
	defer closeFiles()

	imported, err := s.service.Import(files)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_csv", err.Error())
		return
	}

	models := make([]transactionModel, 0, len(imported))
	for _, item := range imported {
		models = append(models, transactionToModel(item.ID, item.Transaction))
	}
	writeJSON(w, http.StatusOK, envelope{Data: importResponse{
		FilesCount:   len(fileHeaders),
		RowsCount:    len(models),
		Transactions: models,
	}})
}

func (s *Server) calculateSplit(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var request calculationRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if len(request.Transactions) == 0 {
		writeError(w, http.StatusBadRequest, "missing_transactions", "at least one transaction is required")
		return
	}

	input := application.AnalysisInput{
		Prefixes:     request.Prefixes,
		Transactions: make([]application.AnalysisTransaction, 0, len(request.Transactions)),
	}
	for index, model := range request.Transactions {
		tx, err := model.toDomain()
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_transaction", fmt.Sprintf("transaction %d: %v", index, err))
			return
		}
		input.Transactions = append(input.Transactions, application.AnalysisTransaction{
			ID:               model.ID,
			Transaction:      tx,
			Selection:        model.Selection,
			SplitAmountCents: model.SplitAmountCents,
			Allocation:       model.Allocation,
		})
	}

	result, err := s.service.Analyze(input)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_calculation", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, envelope{Data: calculationResponse{
		Included:  rowsToModels(result.Included),
		Unmatched: rowsToModels(result.Unmatched),
		Totals: totalsModel{
			TotalCents:              result.Totals.TotalCents,
			ParticipantOneHalfCents: result.Totals.ParticipantOneHalfCents,
			ParticipantTwoHalfCents: result.Totals.ParticipantTwoHalfCents,
		},
	}})
}

func openImportFiles(headers []*multipart.FileHeader) ([]application.ImportFile, func(), error) {
	opened := make([]multipart.File, 0, len(headers))
	files := make([]application.ImportFile, 0, len(headers))
	closeFiles := func() {
		for _, file := range opened {
			_ = file.Close()
		}
	}

	for _, header := range headers {
		file, err := header.Open()
		if err != nil {
			closeFiles()
			return nil, func() {}, fmt.Errorf("%s: open uploaded CSV: %w", header.Filename, err)
		}
		opened = append(opened, file)
		files = append(files, application.ImportFile{Name: header.Filename, Reader: file})
	}
	return files, closeFiles, nil
}

func (m transactionModel) toDomain() (transaction.Transaction, error) {
	date, err := time.Parse("2006-01-02", strings.TrimSpace(m.Date))
	if err != nil {
		return transaction.Transaction{}, fmt.Errorf("date must use YYYY-MM-DD")
	}
	if strings.TrimSpace(m.ID) == "" {
		return transaction.Transaction{}, fmt.Errorf("id is required")
	}
	if strings.TrimSpace(m.Description) == "" {
		return transaction.Transaction{}, fmt.Errorf("description is required")
	}
	if m.SourceLine < 0 {
		return transaction.Transaction{}, fmt.Errorf("source_line cannot be negative")
	}
	return transaction.Transaction{
		Date:        date,
		Description: strings.TrimSpace(m.Description),
		AmountCents: m.AmountCents,
		SourceFile:  m.SourceFile,
		SourceLine:  m.SourceLine,
	}, nil
}

func transactionToModel(id string, tx transaction.Transaction) transactionModel {
	return transactionModel{
		ID:          id,
		Date:        tx.Date.Format("2006-01-02"),
		Description: tx.Description,
		AmountCents: tx.AmountCents,
		SourceFile:  tx.SourceFile,
		SourceLine:  tx.SourceLine,
	}
}

func rowsToModels(rows []application.AnalysisRow) []calculationRow {
	models := make([]calculationRow, 0, len(rows))
	for _, row := range rows {
		models = append(models, calculationRow{
			transactionModel: transactionToModel(row.ID, row.Transaction),
			SplitAmountCents: row.SplitAmountCents,
			Allocation:       row.Allocation,
		})
	}
	return models
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("request body must be valid JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("request body must contain one JSON value")
	}
	return nil
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", fmt.Sprintf("use %s for this endpoint", method))
	return false
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, envelope{Error: &apiError{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, value envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(s.config.AllowedOrigins))
	allowAny := false
	for _, origin := range s.config.AllowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin == "*" {
			allowAny = true
		} else if origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Add("Vary", "Origin")
			if allowAny {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
