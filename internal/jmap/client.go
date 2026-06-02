package jmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	BaseURL   string
	Username  string
	Password  string
	AccountID string
	Session   *Session
	HTTP      *http.Client
}

func NewClient(baseURL string, user string, pass string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		BaseURL:  baseURL,
		Username: user,
		Password: pass,

		HTTP: &http.Client{},
	}
}

func (c *Client) Connect() error {
	req, err := c.NewRequest("GET", c.AuthURL(), nil)

	if err != nil {
		return err
	}

	resp, err := c.HTTP.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
	}

	var session Session

	err = json.Unmarshal(data, &session)

	if err != nil {
		return err
	}

	c.Session = &session

	accountID := session.PrimaryAccounts["urn:ietf:params:jmap:mail"]

	// fallback for your Stalwart setup
	if accountID == "" {
		return fmt.Errorf("mail account id not found in JMAP session")
	}

	c.AccountID = accountID

	return nil
}

func (c *Client) APIURL() string {
	return c.BaseURL + "/"
}

func (c *Client) AuthURL() string {
	return c.BaseURL + "/session"
}

func (c *Client) UploadURL() string {
	return c.BaseURL + "/upload/" + c.AccountID + "/"
}

func (c *Client) DownloadURL(blobID string, name string) string {
	return c.BaseURL + "/download/" + c.AccountID + "/" + blobID + "/" + name
}

func (c *Client) NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Username, c.Password)

	return req, nil
}

func (c *Client) Call(body any) ([]byte, error) {
	b, err := json.Marshal(body)

	if err != nil {
		return nil, err
	}

	req, err := c.NewRequest("POST", c.APIURL(), bytes.NewBuffer(b))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}
