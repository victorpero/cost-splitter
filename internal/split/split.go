package split

import (
	"fmt"
	"strings"

	"github.com/victorpero/amex-grocery-splitter-se/internal/transaction"
)

// Allocation identifies how one transaction is assigned between the two
// participants.
type Allocation string

const (
	AllocationSplitEvenly    Allocation = "split_evenly"
	AllocationParticipantOne Allocation = "participant_one"
	AllocationParticipantTwo Allocation = "participant_two"
)

// AllocatedTransaction couples a transaction with its own allocation. Each row
// is calculated independently, so changing one allocation never affects the
// allocation of another row.
type AllocatedTransaction struct {
	Transaction      transaction.Transaction
	SplitAmountCents int64
	Allocation       Allocation
}

type Result struct {
	TotalCents              int64
	ParticipantOneHalfCents int64
	ParticipantTwoHalfCents int64
}

func ParseAllocation(value string) (Allocation, error) {
	allocation := Allocation(strings.TrimSpace(value))
	switch allocation {
	case AllocationSplitEvenly, AllocationParticipantOne, AllocationParticipantTwo:
		return allocation, nil
	default:
		return "", fmt.Errorf("invalid allocation %q", value)
	}
}

// Calculate preserves the historical CLI behavior by splitting every supplied
// transaction evenly between the two participants.
func Calculate(transactions []transaction.Transaction) Result {
	allocated := make([]AllocatedTransaction, 0, len(transactions))
	for _, tx := range transactions {
		allocated = append(allocated, AllocatedTransaction{
			Transaction:      tx,
			SplitAmountCents: tx.AmountCents,
			Allocation:       AllocationSplitEvenly,
		})
	}
	return CalculateAllocated(allocated)
}

// CalculateAllocated totals signed transaction values and applies each row's
// selected allocation. Participant amounts use half-cent units so an odd-cent
// transaction can still be divided exactly when split evenly.
func CalculateAllocated(transactions []AllocatedTransaction) Result {
	var result Result
	for _, allocated := range transactions {
		amount := allocated.SplitAmountCents
		result.TotalCents += amount

		switch allocated.Allocation {
		case AllocationParticipantOne:
			result.ParticipantOneHalfCents += amount * 2
		case AllocationParticipantTwo:
			result.ParticipantTwoHalfCents += amount * 2
		default:
			result.ParticipantOneHalfCents += amount
			result.ParticipantTwoHalfCents += amount
		}
	}
	return result
}

func FormatCents(currency string, cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}

	whole := cents / 100
	fraction := cents % 100
	return fmt.Sprintf("%s%s %s,%02d", sign, strings.TrimSpace(currency), formatWhole(whole), fraction)
}

// FormatHalfCents formats an amount expressed in half-cent units.
func FormatHalfCents(currency string, halfCents int64) string {
	sign := ""
	if halfCents < 0 {
		sign = "-"
		halfCents = -halfCents
	}

	thousandths := halfCents * 5
	whole := thousandths / 1000
	fraction := thousandths % 1000
	if fraction%10 == 0 {
		return fmt.Sprintf("%s%s %s,%02d", sign, strings.TrimSpace(currency), formatWhole(whole), fraction/10)
	}
	return fmt.Sprintf("%s%s %s,%03d", sign, strings.TrimSpace(currency), formatWhole(whole), fraction)
}

func formatWhole(value int64) string {
	text := fmt.Sprintf("%d", value)
	if len(text) <= 3 {
		return text
	}

	var builder strings.Builder
	firstGroup := len(text) % 3
	if firstGroup == 0 {
		firstGroup = 3
	}
	builder.WriteString(text[:firstGroup])
	for i := firstGroup; i < len(text); i += 3 {
		builder.WriteByte(' ')
		builder.WriteString(text[i : i+3])
	}
	return builder.String()
}
