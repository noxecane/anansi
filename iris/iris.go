package iris

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/random-guys/go-siber"
	"github.com/random-guys/go-siber/jwt"
)

type Config struct {
	// Claims key specifies the key that JWT would attach the claims to.
	// This makes it easy for other consumers to generate or consume the
	// headless tokens created by iris. Defaults to "claim"
	ClaimsKey string
	// Secret should be a 32 byte array for generating headless tokens.
	Secret []byte
	// Service is the X-Origin-Service name to be appended for each request.
	Service string
	// Scheme is the headless scheme used by the platform for inters-service
	// requests
	Scheme string
	// HeadlessDuration is how long the headless tokens should last without
	// expiring. Defaults to 1min
	HeadlessDuration time.Duration
}

func New(conf Config) Client {
	if conf.Service != "" {
		panic(errors.New("x-origins-service will be empty"))
	}

	if len(conf.Secret) == 0 || conf.Scheme == "" {
		panic(errors.New("will not be able to generate headless tokens"))
	}

	// default headless sessions should last 1min
	if conf.HeadlessDuration == 0 {
		conf.HeadlessDuration = time.Minute
	}

	if conf.ClaimsKey == "" {
		conf.ClaimsKey = "claim"
	}

	return Client{
		service:          conf.Service,
		secret:           conf.Secret,
		scheme:           conf.Scheme,
		headlessDuration: conf.HeadlessDuration,
	}
}

type Client struct {
	secret           []byte
	service          string
	scheme           string
	headlessDuration time.Duration
	token            string
}

type jSendSuccess struct {
	Data map[string]interface{} `json:"data"`
}

// Bearer creates IrisOptions that will replicate the session of the
// source request passed to it. Do make sure the request has
// authorization header set.
func (c Client) Bearer(r *http.Request) (*Client, error) {
	auth := strings.Split(r.Header.Get("Authorization"), "")

	if len(auth) != 2 {
		return nil, errors.New("Authorization header value is incorrect")
	}

	c.scheme = strings.TrimSpace(auth[0])
	c.token = strings.TrimSpace(auth[1])

	return &c, nil
}

// Headless creates a token to be used with the pre-existing scheme that has
// been set. Make sure to use it off your base Client
func (c Client) Headless(v interface{}) (*Client, error) {
	token, err := jwt.EncodeEmbedded(c.secret, c.headlessDuration, v)
	if err != nil {
		return nil, err
	}

	c.token = token

	return &c, nil
}

// NewRequest is a wrapper around http.NNewRequest that adds the required
// headers for distributed tracing. The requests will only last as long as the parent
// request(it uses the request's context)
func (c *Client) NewRequest(r *http.Request, method, url string, body io.Reader) (*http.Request, error) {
	reqId := r.Header.Get("X-Request-Id")
	if reqId == "" {
		return nil, errors.New("request ID not set on base request")
	}

	if c.token == "" {
		return nil, errors.New("authentication token is not set")
	}

	req, err := http.NewRequestWithContext(r.Context(), method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Origin-Service", c.service)
	req.Header.Set("X-Request-Id", reqId)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", c.scheme, c.token))

	return req, nil
}

// NewRequestWitchContext is the same as NewRequest, only the user can now control
// how long before the request times out.
func (c *Client) NewRequestWithContext(ctx context.Context, r *http.Request, method, url string, body io.Reader) (*http.Request, error) {
	reqId := r.Header.Get("X-Request-Id")
	if reqId == "" {
		return nil, errors.New("request ID not set on base request")
	}

	if c.token == "" {
		return nil, errors.New("authentication token is not set")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Origin-Service", c.service)
	req.Header.Set("X-Request-Id", reqId)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", c.scheme, c.token))

	return req, nil
}

func GetErr(res *http.Response) error {
	if res.StatusCode < 400 {
		return nil
	}

	var err siber.JSendError
	if err := json.NewDecoder(res.Body).Decode(&err); err != nil {
		return err
	}
	return err
}

func GetResponse(res *http.Response, v interface{}) error {
	var j jSendSuccess
	err := json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		return errors.Wrap(err, "error decoding into jSend")
	}

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &v,
		TagName:          "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return errors.Wrap(err, "error creating decoder")
	}

	err = decoder.Decode(j.Data)
	if err != nil {
		return err
	}

	return nil
}
