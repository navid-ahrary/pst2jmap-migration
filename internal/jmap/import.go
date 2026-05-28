package jmap

func (c *Client) ImportEmail(
	blobID string,
	mailboxID string,
) error {

	body :=
		map[string]any{
			"using": []string{
				"urn:ietf:params:jmap:core",
				"urn:ietf:params:jmap:mail",
			},

			"methodCalls": [][]any{
				{
					"Email/import",

					map[string]any{
						"accountId": c.AccountID,

						"emails": map[string]any{
							"1": map[string]any{
								"blobId": blobID,

								"mailboxIds": map[string]bool{
									mailboxID: true,
								},
							},
						},
					},

					"0",
				},
			},
		}

	_, err :=
		c.Call(body)

	return err
}
