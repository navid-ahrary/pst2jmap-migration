package jmap

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetMailboxIDs() (
	map[string]string,
	error,
) {

	body := map[string]any{
		"using": []string{
			"urn:ietf:params:jmap:core",
			"urn:ietf:params:jmap:mail",
		},

		"methodCalls": [][]any{
			{
				"Mailbox/get",

				map[string]any{
					"accountId": c.AccountID,
				},

				"0",
			},
		},
	}

	data, err := c.Call(body)

	if err != nil {
		return nil, err
	}

	var result struct {
		MethodResponses [][]json.RawMessage `json:"methodResponses"`
	}

	err = json.Unmarshal(
		data,
		&result,
	)

	if err != nil {
		return nil, err
	}

	if len(result.MethodResponses) == 0 {
		return nil, fmt.Errorf("empty mailbox response")
	}

	var response struct {
		List []struct {
			ID   string `json:"id"`
			Role string `json:"role"`
			Name string `json:"name"`
		} `json:"list"`
	}

	err = json.Unmarshal(
		result.MethodResponses[0][1],
		&response,
	)

	if err != nil {
		return nil, err
	}

	mailboxes := map[string]string{}

	for _, mb := range response.List {

		if mb.Role == "" {
			continue
		}

		mailboxes[mb.Role] = mb.ID
	}

	if mailboxes["inbox"] == "" {
		return nil, fmt.Errorf(
			"inbox mailbox not found",
		)
	}

	return mailboxes, nil
}

func ResolveMailboxID(
	folder string,
	mailboxes map[string]string,
) string {

	switch folder {

	case "Inbox":
		if id := mailboxes["inbox"]; id != "" {
			return id
		}

	case "Sent Items", "Sent-Items":
		if id := mailboxes["sent"]; id != "" {
			return id
		}

	case "Deleted Items", "Deleted-Items":
		if id := mailboxes["trash"]; id != "" {
			return id
		}

	case "Drafts":
		if id := mailboxes["drafts"]; id != "" {
			return id
		}

	case "Junk Email", "Junk-Email":
		if id := mailboxes["junk"]; id != "" {
			return id
		}
	}

	return mailboxes["inbox"]
}
