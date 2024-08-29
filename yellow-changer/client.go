package yellowChanger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/nlypage/yellow-changer/yellow-changer/common"
	"io"
	"net/http"
	"time"
)

const (
	apiUrl = "https://api.yellowchanger.com/"
)

// Client is a struct to interact with the yellow-changer API.
type Client struct {
	publicKey  string
	privateKey string

	httpClient *http.Client
}

type Options struct {
	// PublicKey from your YellowChanger personal cabinet.
	PublicKey string
	// PrivateKey from your YellowChanger personal cabinet.
	PrivateKey string
	// ClientTimeout field is optional. Default value is 30 seconds.
	ClientTimeout time.Duration
}

// NewClient creates a new YellowChanger client to interact with the API.
func NewClient(options Options) *Client {
	c := &Client{
		publicKey:  options.PublicKey,
		privateKey: options.PrivateKey,
	}
	clientTimeout := 30 * time.Second
	if options.ClientTimeout != 0 {
		clientTimeout = options.ClientTimeout
	}

	c.httpClient = &http.Client{
		Timeout: clientTimeout,
	}
	return c
}

// Do send a request to the YellowChanger API and returns the response body.
func (c *Client) Do(ctx context.Context, r *Request) ([]byte, error) {
	url := apiUrl + r.Endpoint

	jsonBody, errMarshal := json.Marshal(r.Body)
	if errMarshal != nil {
		return nil, errMarshal
	}

	var (
		req              *http.Request
		errCreateRequest error
	)
	if r.Body != nil {
		req, errCreateRequest = http.NewRequest(r.Method, url, bytes.NewBuffer(jsonBody))
		if errCreateRequest != nil {
			return nil, errCreateRequest
		}
	} else {
		req, errCreateRequest = http.NewRequest(r.Method, url, nil)
		if errCreateRequest != nil {
			return nil, errCreateRequest
		}
	}

	req = req.WithContext(ctx)

	req.Header.Set("Y_API_KEY", c.publicKey)

	if r.Body != nil {
		signature := common.GenerateHmacSignature(string(jsonBody), c.privateKey)
		req.Header.Set("Signature", signature)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errorResponse ErrorResponse
		if errUnmarshall := json.Unmarshal(responseBody, &errorResponse); errUnmarshall != nil {
			return nil, errUnmarshall
		}
		return nil, fmt.Errorf("error response with code %d from the api: %s", errorResponse.StatusCode, errorResponse.Message)
	}

	return responseBody, nil
}
