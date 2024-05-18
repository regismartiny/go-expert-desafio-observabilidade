package viacep

import (
	"context"
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
	if c.apiKey != "" {
		req.Header.Set("Authentication", fmt.Sprintf("bearer %s", c.apiKey))
	}

	res, err := http.DefaultClient.Do(req)
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

func (c *Client) GetTemperatureInfo(ctx *context.Context, cep string) (TemperatureInfo, error) {

	url := fmt.Sprintf("%s/%s", c.BaseURL, cep)

	req, err := http.NewRequestWithContext(*ctx, "GET", url, nil)
	if err != nil {
		return TemperatureInfo{}, err
	}

	var res TemperatureInfo
	if err := c.sendRequest(req, &res); err != nil {
		return TemperatureInfo{}, err
	}

	return res, nil
}
