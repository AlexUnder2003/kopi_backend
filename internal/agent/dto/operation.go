package dto

type OperationItem struct {
	Name           string `json:"name"`
	Amount         string `json:"amount"`
	Type           string `json:"type"`
	Account        string `json:"account"`
	Currency       string `json:"currency"`
	FromAccount    string `json:"from_account,omitempty"`
	Category       string `json:"category"`
	OccurrenceDate string `json:"occurrence_date"`
}

type ListOperationsArgs struct {
	Filter     string `json:"filter,omitempty" jsonschema:"Optional operation name substring."`
	AccountID  string `json:"account_id,omitempty" jsonschema:"Optional account UUID."`
	CategoryID string `json:"category_id,omitempty" jsonschema:"Optional category UUID."`
	StartDate  string `json:"start_date,omitempty" jsonschema:"Optional start date, YYYY-MM-DD."`
	EndDate    string `json:"end_date,omitempty" jsonschema:"Optional end date, YYYY-MM-DD."`
	Limit      int    `json:"limit,omitempty" jsonschema:"Maximum number of rows. Defaults to 50, maximum 100."`
}

type CreateOperationArgs struct {
	Name           string `json:"name" jsonschema:"Short name of the operation, for example coffee or salary."`
	Amount         string `json:"amount" jsonschema:"Positive decimal amount, for example 500 or 12.50."`
	Type           string `json:"type" jsonschema:"income, expense, or transfer."`
	AccountID      string `json:"account_id" jsonschema:"Destination account UUID. For a transfer this is the account that receives money."`
	FromAccountID  string `json:"from_account_id,omitempty" jsonschema:"Source account UUID. Required for transfer."`
	CategoryID     string `json:"category_id" jsonschema:"Category UUID."`
	OccurrenceDate string `json:"occurrence_date,omitempty" jsonschema:"When it happened, YYYY-MM-DD. Defaults to now."`
}
