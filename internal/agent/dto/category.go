package dto

type CategoryItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsSystem bool   `json:"is_system"`
}

type ListCategoriesArgs struct {
	Filter string `json:"filter,omitempty" jsonschema:"Optional category name substring."`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum number of rows. Defaults to 50, maximum 100."`
}
