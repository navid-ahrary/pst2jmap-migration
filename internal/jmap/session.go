package jmap

type Session struct {
	Capabilities map[string]any `json:"capabilities"`

	Accounts map[string]any `json:"accounts"`

	PrimaryAccounts map[string]string `json:"primaryAccounts"`

	Username string `json:"username"`

	APIURL string `json:"apiUrl"`

	DownloadURL string `json:"downloadUrl"`

	UploadURL string `json:"uploadUrl"`

	EventSourceURL string `json:"eventSourceUrl"`

	State string `json:"state"`
}
