package dto

type PlannedOperationItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Amount       string `json:"amount"`
	Type         string `json:"type"`
	AccountID    string `json:"account_id"`
	Account      string `json:"account"`
	Currency     string `json:"currency"`
	CategoryID   string `json:"category_id"`
	Category     string `json:"category"`
	IntervalType string `json:"interval_type"`
	PlannedDate  string `json:"planned_date"`
	NextRunDate  string `json:"next_run_date,omitempty"`
	IsRecurring  bool   `json:"is_recurring"`
}

type ListPlannedOperationsArgs struct {
	Filter    string `json:"filter,omitempty" jsonschema:"Optional planned operation name substring."`
	StartDate string `json:"start_date,omitempty" jsonschema:"Optional start of the next run window, YYYY-MM-DD."`
	EndDate   string `json:"end_date,omitempty" jsonschema:"Optional end of the next run window, YYYY-MM-DD."`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of rows. Defaults to 50, maximum 100."`
}

type CreatePlannedOperationArgs struct {
	Name         string `json:"name" jsonschema:"Short name, for example rent."`
	Amount       string `json:"amount" jsonschema:"Positive decimal amount."`
	Type         string `json:"type" jsonschema:"income or expense."`
	AccountID    string `json:"account_id" jsonschema:"Account UUID from list_accounts."`
	CategoryID   string `json:"category_id" jsonschema:"Category UUID from list_categories."`
	IntervalType string `json:"interval_type" jsonschema:"daily, weekly, biweekly, monthly, or custom."`
	Interval     int    `json:"interval,omitempty" jsonschema:"Number of days. Required when interval_type is custom."`
	PlannedDate  string `json:"planned_date" jsonschema:"First run date, YYYY-MM-DD."`
	IsRecurring  bool   `json:"is_recurring,omitempty" jsonschema:"Repeat on the interval. Defaults to false."`
}
