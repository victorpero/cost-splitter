package transaction

import "time"

// Transaction is the normalized representation used by the rest of the app.
// AmountCents keeps the sign found in the CSV throughout parsing, storage, and
// split calculations.
type Transaction struct {
	Date        time.Time
	Description string
	AmountCents int64
	SourceFile  string
	SourceLine  int
}
