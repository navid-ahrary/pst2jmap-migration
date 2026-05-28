package jmap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func (c *Client) UploadEML(
	path string,
) (string, error) {

	f, err := os.Open(path)

	if err != nil {
		return "", err
	}

	defer f.Close()

	url := c.UploadURL()

	req, err := http.NewRequest(
		"POST",
		url,
		f,
	)

	if err != nil {
		return "", err
	}

	req.SetBasicAuth(
		c.Username,
		c.Password,
	)

	req.Header.Set(
		"Content-Type",
		"message/rfc822",
	)

	resp, err := c.HTTP.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(
		resp.Body,
	)

	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 300 {

		return "", fmt.Errorf(
			"upload failed: %s",
			string(data),
		)
	}

	var result struct {
		BlobID string `json:"blobId"`
	}

	err = json.Unmarshal(
		data,
		&result,
	)

	return result.BlobID, err
}
