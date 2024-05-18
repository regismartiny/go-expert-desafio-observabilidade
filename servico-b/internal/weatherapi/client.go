package weatherapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

var (
	APPLICATION_JSON_UTF8 = "application/json; charset=utf-8"
)

type Client struct {
	BaseURL *url.URL
	apiKey  string
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewClient(baseUrl *url.URL, apiKey string) *Client {
	return &Client{
		BaseURL: baseUrl,
		apiKey:  apiKey,
	}
}

func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", APPLICATION_JSON_UTF8)
	req.Header.Set("Accept", APPLICATION_JSON_UTF8)

	// Insecure client
	// Required due to error: tls: failed to verify certificate: x509: certificate signed by unknown authority
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: customTransport}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return errors.New(errRes.Message)
		}

		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	if err = json.NewDecoder(res.Body).Decode(v); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetWeatherInfo(ctx *context.Context, cidade string) (Weather, error) {

	url := c.BaseURL.JoinPath("/current.json")

	q := url.Query()
	q.Set("key", c.apiKey)
	q.Set("aqi", "no")
	q.Set("q", cidade)
	url.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(*ctx, "GET", url.String(), nil)
	if err != nil {
		return Weather{}, err
	}

	var res Weather
	if err := c.sendRequest(req, &res); err != nil {
		return Weather{}, err
	}

	return res, nil
}
