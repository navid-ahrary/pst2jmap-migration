package jmap

type Session struct {
	Accounts map[string]any `json:"accounts"`

	PrimaryAccounts map[string]string `json:"primaryAccounts"`

	Username string `json:"username"`

	State string `json:"state"`
}
