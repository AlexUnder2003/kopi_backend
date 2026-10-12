package dto

type AccountItem struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Currency             string `json:"currency"`
	Balance              string `json:"balance"`
	IncludeInFreeBalance bool   `json:"include_in_free_balance"`
}

type ListAccountsArgs struct {
	Currency string `json:"currency" jsonschema:"ISO currency code used to convert the total balance, for example USD or RUB."`
	Filter   string `json:"filter,omitempty" jsonschema:"Optional account name substring."`
	Limit    int    `json:"limit,omitempty" jsonschema:"Maximum number of rows. Defaults to 50, maximum 100."`
}
