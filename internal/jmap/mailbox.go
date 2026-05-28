package jmap

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetInboxID() (
	string,
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
		return "", err
	}

	var result struct {
		MethodResponses [][]json.RawMessage `json:"methodResponses"`
	}

	err = json.Unmarshal(
		data,
		&result,
	)

	if err != nil {
		return "", err
	}

	var response struct {
		List []struct {
			ID   string `json:"id"`
			Role string `json:"role"`
			Name string `json:"name"`
		} `json:"list"`
	}

	err =
		json.Unmarshal(
			result.MethodResponses[0][1],
			&response,
		)

	if err != nil {
		return "", err
	}

	for _, mb := range response.List {

		if mb.Role == "inbox" {
			return mb.ID, nil
		}
	}

	return "", fmt.Errorf(
		"inbox mailbox not found",
	)
}
