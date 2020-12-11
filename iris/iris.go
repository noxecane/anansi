package iris

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/random-guys/go-siber/jsend"
	"github.com/random-guys/go-siber/jwt"
)

var (
	// ErrNoRequestID is returned when the parent request doesn't have a request ID
	ErrNoRequestID = errors.New("no request id")
	// ErrNoAuthentication is returned when there's no authentication information
	// attached to the parent request
	ErrNoAuthentication = errors.New("no authentication")
)

type Config struct {
	// Secret should be a 32 byte array for generating headless tokens.
	Secret []byte
	// Service is the X-Origin-Service name to be appended for each request.
	Service string
	// HeadlessScheme is the headless scheme used by the platform for inters-service requests
	HeadlessScheme string
	// HeadlessDuration is how long the headless tokens should last without
	// expiring. Defaults to 1min
	HeadlessDuration time.Duration
}

func NewClient(conf Config) Client {
	if conf.Service == "" {
		panic(errors.New("x-origins-service will be empty"))
	}

	if len(conf.Secret) == 0 || conf.HeadlessScheme == "" {
		panic(errors.New("will not be able to generate headless tokens"))
	}

	// default headless sessions should last 1min
	if conf.HeadlessDuration == 0 {
		conf.HeadlessDuration = time.Minute
	}

	return Client{
		serviceName:      conf.Service,
		serviceSecret:    conf.Secret,
		headlessScheme:   conf.HeadlessScheme,
		headlessDuration: conf.HeadlessDuration,
	}
}

type Client struct {
	serviceSecret    []byte
	serviceName      string
	headlessScheme   string
	headlessDuration time.Duration
}

type jSendSuccess struct {
	Data map[string]interface{} `json:"data"`
}

// NewRequest is a wrapper around http.NewRequest that proxies the authentication and enables
// distributed tracing. Note this request lasts as long as the parent request.
func (c *Client) NewRequest(r *http.Request, method, url string, body io.Reader) (*http.Request, error) {
	reqId := r.Header.Get("X-Request-Id")
	if reqId == "" {
		return nil, ErrNoRequestID
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, ErrNoAuthentication
	}

	req, err := http.NewRequestWithContext(r.Context(), method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Origin-Service", c.serviceName)
	req.Header.Set("X-Request-Id", reqId)
	req.Header.Set("Authorization", auth)

	return req, nil
}

// NewHeadlessRequest is like NewRequest but it replaces the request authentication mechanism with
// an headless one, giving users control over what would pass for the session on the receiving server.
func (c *Client) NewHeadlessRequest(r *http.Request, method, url string, session interface{}, body io.Reader) (*http.Request, error) {
	reqId := r.Header.Get("X-Request-Id")
	if reqId == "" {
		return nil, ErrNoRequestID
	}

	token, err := jwt.EncodeEmbedded(c.serviceSecret, c.headlessDuration, session)
	if err != nil {
		return nil, errors.Wrap(err, "could not create headless token")
	}

	req, err := http.NewRequestWithContext(r.Context(), method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Origin-Service", c.serviceName)
	req.Header.Set("X-Request-Id", reqId)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", c.headlessScheme, token))

	return req, nil
}

// NewHeadlessRequest is like NewRequest but it replaces the request authentication mechanism with
// an headless one, giving users control over what would pass for the session on the receiving server.
func (c *Client) NewBaseRequest(ctx context.Context, method, url string, session interface{}, body io.Reader) (*http.Request, error) {
	reqId := NextRequestID()

	token, err := jwt.EncodeEmbedded(c.serviceSecret, c.headlessDuration, session)
	if err != nil {
		return nil, errors.Wrap(err, "could not create headless token")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Origin-Service", c.serviceName)
	req.Header.Set("X-Request-Id", reqId)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", c.headlessScheme, token))

	return req, nil
}

func GetErr(res *http.Response) error {
	if res.StatusCode < 400 {
		return nil
	}

	var err jsend.Err
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
