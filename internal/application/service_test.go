package application

import (
	"strings"
	"testing"
	"time"

	"github.com/victorpero/cost-splitter/internal/split"
	"github.com/victorpero/cost-splitter/internal/transaction"
)

func TestServiceImportAndAnalyzePreservesCostSplittingBehavior(t *testing.T) {
	service := NewService()
	imported, err := service.Import([]ImportFile{{
		Name:   "activity.csv",
		Reader: strings.NewReader("Datum;Beskrivning;Belopp\n2026-05-11;COOP RADHUSET;-111,00\n2026-05-12;APOTEKET;50,00\n"),
	}})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if len(imported) != 2 {
		t.Fatalf("len(imported) = %d, want 2", len(imported))
	}

	adjustedAmount := int64(-6100)
	result, err := service.Analyze(AnalysisInput{
		Prefixes: []string{"ICA", "COOP"},
		Transactions: []AnalysisTransaction{
			{
				ID:               imported[0].ID,
				Transaction:      imported[0].Transaction,
				SplitAmountCents: &adjustedAmount,
				Allocation:       split.AllocationParticipantOne,
			},
			{
				ID:          imported[1].ID,
				Transaction: imported[1].Transaction,
				Selection:   SelectionIncluded,
				Allocation:  split.AllocationParticipantTwo,
			},
		},
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(result.Included) != 2 || len(result.Unmatched) != 0 {
		t.Fatalf("included/unmatched = %d/%d, want 2/0", len(result.Included), len(result.Unmatched))
	}
	if result.Totals.TotalCents != -1100 || result.Totals.ParticipantOneHalfCents != -12200 || result.Totals.ParticipantTwoHalfCents != 10000 {
		t.Fatalf("totals = %+v", result.Totals)
	}
}

func TestServiceAnalyzeSupportsExplicitExclusionAndSwedishMatching(t *testing.T) {
	service := NewService()
	date := time.Date(2026, time.May, 11, 0, 0, 0, 0, time.UTC)
	result, err := service.Analyze(AnalysisInput{
		Prefixes: []string{"PRESSBYRÅN"},
		Transactions: []AnalysisTransaction{
			{ID: "one", Transaction: transaction.Transaction{Date: date, Description: "Pressbyran Central", AmountCents: -1000}},
			{ID: "two", Transaction: transaction.Transaction{Date: date, Description: "PRESSBYRÅN CITY", AmountCents: -2000}, Selection: SelectionExcluded},
		},
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(result.Included) != 1 || result.Included[0].ID != "one" {
		t.Fatalf("included = %+v", result.Included)
	}
	if len(result.Unmatched) != 1 || result.Unmatched[0].ID != "two" {
		t.Fatalf("unmatched = %+v", result.Unmatched)
	}
}

func TestServiceAnalyzeRejectsInvalidReusableContractInput(t *testing.T) {
	service := NewService()
	date := time.Date(2026, time.May, 11, 0, 0, 0, 0, time.UTC)
	_, err := service.Analyze(AnalysisInput{
		Prefixes: []string{"ICA"},
		Transactions: []AnalysisTransaction{
			{ID: "duplicate", Transaction: transaction.Transaction{Date: date}, Selection: SelectionAutomatic},
			{ID: "duplicate", Transaction: transaction.Transaction{Date: date}, Selection: SelectionAutomatic},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("Analyze() error = %v, want duplicate id error", err)
	}
}
