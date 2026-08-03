package split

import (
	"testing"

	"github.com/victorpero/amex-grocery-splitter-se/internal/transaction"
)

func TestCalculatePreservesMixedTransactionSigns(t *testing.T) {
	transactions := []transaction.Transaction{
		{AmountCents: -10050},
		{AmountCents: 2500},
	}

	result := Calculate(transactions)
	if result.TotalCents != -7550 {
		t.Fatalf("TotalCents = %d, want -7550", result.TotalCents)
	}
	if result.ParticipantOneHalfCents != -7550 {
		t.Fatalf("ParticipantOneHalfCents = %d, want -7550", result.ParticipantOneHalfCents)
	}
	if result.ParticipantTwoHalfCents != -7550 {
		t.Fatalf("ParticipantTwoHalfCents = %d, want -7550", result.ParticipantTwoHalfCents)
	}
}

func TestCalculateAllocatedKeepsEachTransactionIndependent(t *testing.T) {
	transactions := []AllocatedTransaction{
		{
			Transaction: transaction.Transaction{AmountCents: -10050},
			Allocation:  AllocationParticipantOne,
		},
		{
			Transaction: transaction.Transaction{AmountCents: 2500},
			Allocation:  AllocationParticipantTwo,
		},
		{
			Transaction: transaction.Transaction{AmountCents: -3000},
			Allocation:  AllocationSplitEvenly,
		},
	}

	result := CalculateAllocated(transactions)
	if result.TotalCents != -10550 {
		t.Fatalf("TotalCents = %d, want -10550", result.TotalCents)
	}
	if result.ParticipantOneHalfCents != -23100 {
		t.Fatalf("ParticipantOneHalfCents = %d, want -23100", result.ParticipantOneHalfCents)
	}
	if result.ParticipantTwoHalfCents != 2000 {
		t.Fatalf("ParticipantTwoHalfCents = %d, want 2000", result.ParticipantTwoHalfCents)
	}
}

func TestChangingOneAllocationDoesNotChangeAnother(t *testing.T) {
	transactions := []AllocatedTransaction{
		{Transaction: transaction.Transaction{AmountCents: 1000}, Allocation: AllocationSplitEvenly},
		{Transaction: transaction.Transaction{AmountCents: 2500}, Allocation: AllocationParticipantTwo},
	}

	before := CalculateAllocated(transactions)
	transactions[0].Allocation = AllocationParticipantOne
	after := CalculateAllocated(transactions)

	if before.ParticipantTwoHalfCents != 6000 {
		t.Fatalf("before ParticipantTwoHalfCents = %d, want 6000", before.ParticipantTwoHalfCents)
	}
	if after.ParticipantTwoHalfCents != 5000 {
		t.Fatalf("after ParticipantTwoHalfCents = %d, want 5000", after.ParticipantTwoHalfCents)
	}
	if transactions[1].Allocation != AllocationParticipantTwo {
		t.Fatalf("second allocation = %q, want %q", transactions[1].Allocation, AllocationParticipantTwo)
	}
}

func TestFormatHalfCentsShowsHalfCentWhenNeeded(t *testing.T) {
	got := FormatHalfCents("SEK", 10001)
	want := "SEK 50,005"
	if got != want {
		t.Fatalf("FormatHalfCents = %q, want %q", got, want)
	}
}

func TestFormatCentsUsesSwedishStyleSeparators(t *testing.T) {
	got := FormatCents("SEK", 1234567)
	want := "SEK 12 345,67"
	if got != want {
		t.Fatalf("FormatCents = %q, want %q", got, want)
	}
}
