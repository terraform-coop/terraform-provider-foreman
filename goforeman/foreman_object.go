package goforeman

import "encoding/json"

type ForemanObject struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type QueryResponse struct {
	Total    int               `json:"total"`
	Subtotal int               `json:"subtotal"`
	Page     int               `json:"page"`
	PerPage  int               `json:"per_page"`
	Search   string            `json:"search,omitempty"`
	Sort     QueryResponseSort `json:"sort,omitempty"`
	Results  []json.RawMessage `json:"results"`
}

type QueryResponseSort struct {
	Order string `json:"order,omitempty"`
	By    string `json:"by,omitempty"`
}

type ForemanTask struct {
	ID        int         `json:"id"`
	Pending   bool        `json:"pending"`
	Label     string      `json:"label,omitempty"`
	Result    string      `json:"result,omitempty"`
	Output    interface{} `json:"output,omitempty"`
	Humanized *struct {
		Errors []string `json:"errors,omitempty"`
	} `json:"humanized,omitempty"`
}
