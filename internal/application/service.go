package application

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/victorpero/cost-splitter/internal/matcher"
	"github.com/victorpero/cost-splitter/internal/parser"
	"github.com/victorpero/cost-splitter/internal/split"
	"github.com/victorpero/cost-splitter/internal/transaction"
)

// Service exposes application-level cost splitting operations independently
// from any particular HTTP transport or frontend.
type Service struct{}

type ImportFile struct {
	Name   string
	Reader io.Reader
}

type ImportedTransaction struct {
	ID          string
	Transaction transaction.Transaction
}

type Selection string

const (
	SelectionAutomatic Selection = "automatic"
	SelectionIncluded  Selection = "included"
	SelectionExcluded  Selection = "excluded"
)

type AnalysisTransaction struct {
	ID               string
	Transaction      transaction.Transaction
	Selection        Selection
	SplitAmountCents *int64
	Allocation       split.Allocation
}

type AnalysisInput struct {
	Prefixes     []string
	Transactions []AnalysisTransaction
}

type AnalysisRow struct {
	ID               string
	Transaction      transaction.Transaction
	SplitAmountCents int64
	Allocation       split.Allocation
}

type AnalysisResult struct {
	Included  []AnalysisRow
	Unmatched []AnalysisRow
	Totals    split.Result
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Import(files []ImportFile) ([]ImportedTransaction, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("at least one CSV file is required")
	}

	imported := make([]ImportedTransaction, 0)
	for _, file := range files {
		name := strings.TrimSpace(file.Name)
		if name == "" {
			name = "upload.csv"
		}
		transactions, err := parser.Parse(file.Reader, name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		for _, tx := range transactions {
			imported = append(imported, ImportedTransaction{
				ID:          strconv.Itoa(len(imported)),
				Transaction: tx,
			})
		}
	}
	return imported, nil
}

func (s *Service) Analyze(input AnalysisInput) (AnalysisResult, error) {
	prefixMatcher, err := matcher.NewPrefixMatcher(input.Prefixes)
	if err != nil {
		return AnalysisResult{}, err
	}

	seenIDs := make(map[string]struct{}, len(input.Transactions))
	result := AnalysisResult{
		Included:  make([]AnalysisRow, 0),
		Unmatched: make([]AnalysisRow, 0),
	}

	for _, item := range input.Transactions {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			return AnalysisResult{}, fmt.Errorf("transaction id is required")
		}
		if _, exists := seenIDs[id]; exists {
			return AnalysisResult{}, fmt.Errorf("transaction id %q is duplicated", id)
		}
		seenIDs[id] = struct{}{}

		selection := item.Selection
		if selection == "" {
			selection = SelectionAutomatic
		}
		if err := validateSelection(selection); err != nil {
			return AnalysisResult{}, fmt.Errorf("transaction %s: %w", id, err)
		}

		allocation := item.Allocation
		if allocation == "" {
			allocation = split.AllocationSplitEvenly
		}
		if _, err := split.ParseAllocation(string(allocation)); err != nil {
			return AnalysisResult{}, fmt.Errorf("transaction %s: %w", id, err)
		}

		splitAmount := item.Transaction.AmountCents
		if item.SplitAmountCents != nil {
			splitAmount = *item.SplitAmountCents
		}
		row := AnalysisRow{
			ID:               id,
			Transaction:      item.Transaction,
			SplitAmountCents: splitAmount,
			Allocation:       allocation,
		}

		included := selection == SelectionIncluded ||
			(selection == SelectionAutomatic && prefixMatcher.IsGrocery(item.Transaction.Description))
		if included {
			result.Included = append(result.Included, row)
		} else {
			result.Unmatched = append(result.Unmatched, row)
		}
	}

	sortRows(result.Included)
	sortRows(result.Unmatched)

	allocated := make([]split.AllocatedTransaction, 0, len(result.Included))
	for _, row := range result.Included {
		allocated = append(allocated, split.AllocatedTransaction{
			Transaction:      row.Transaction,
			SplitAmountCents: row.SplitAmountCents,
			Allocation:       row.Allocation,
		})
	}
	result.Totals = split.CalculateAllocated(allocated)
	return result, nil
}

func validateSelection(selection Selection) error {
	switch selection {
	case SelectionAutomatic, SelectionIncluded, SelectionExcluded:
		return nil
	default:
		return fmt.Errorf("invalid selection %q", selection)
	}
}

func sortRows(rows []AnalysisRow) {
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Transaction.Date.Equal(rows[j].Transaction.Date) {
			return rows[i].Transaction.Description < rows[j].Transaction.Description
		}
		return rows[i].Transaction.Date.Before(rows[j].Transaction.Date)
	})
}
