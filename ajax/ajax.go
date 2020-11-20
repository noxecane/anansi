package ajax

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tsaron/anansi"
	"github.com/tsaron/anansi/jwt"
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

type Token struct {
	value    string
	headless bool
}

// BearerToken sets the token to the session token store in the passed request
func (c *Client) BearerToken(r *http.Request) (Token, error) {
	auth := strings.Split(r.Header.Get("Authorization"), " ")

	if len(auth) != 2 {
		return Token{}, fmt.Errorf("authorization header value is incorrect: %s", r.Header.Get("Authorization"))
	}

	return Token{strings.TrimSpace(auth[1]), false}, nil
}

// HeadlessToken creates a token to be used with the client's scheme
func (c *Client) HeadlessToken(v interface{}) (Token, error) {
	token, err := jwt.EncodeStruct(c.serviceSecret, c.headlessDuration, v)
	if err != nil {
		return Token{}, err
	}

	return Token{token, true}, nil
}

// NewRequest is a wrapper around http.NNewRequest that adds the required
// headers for distributed tracing. The requests will only last as long as the parent
// request(it uses the request's context). The request is assigned a random request ID if
// none is found on the request
func (c *Client) NewRequest(r *http.Request, method, url string, token Token, body io.Reader) (*http.Request, error) {
	reqId := r.Header.Get("X-Request-Id")
	if reqId == "" {
		return nil, errors.New("request ID not set on base request")
	}

	req, err := http.NewRequestWithContext(r.Context(), method, url, body)
	if err != nil {
		return nil, err
	}

	var scheme string
	if token.headless {
		scheme = c.headlessScheme
	} else {
		scheme = "Bearer"
	}

	req.Header.Set("X-Origin-Service", c.serviceName)
	req.Header.Set("X-Request-Id", reqId)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", scheme, token.value))

	return req, nil
}

// NewBaseRequest is the same as NewRequest, only the user can now control
// how long before the request times out and it's assigned a random request ID.
// This is best for when a request is being made without a base request.
func (c *Client) NewBaseRequest(ctx context.Context, method, url string, token Token, body io.Reader) (*http.Request, error) {
	reqId := NextRequestID()

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	var scheme string
	if token.headless {
		scheme = c.headlessScheme
	} else {
		scheme = "Bearer"
	}

	req.Header.Set("X-Origin-Service", c.serviceName)
	req.Header.Set("X-Request-Id", reqId)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", scheme, token.value))

	return req, nil
}

func GetErr(res *http.Response) error {
	if res.StatusCode < 400 {
		return nil
	}

	var err anansi.APIError
	if err := json.NewDecoder(res.Body).Decode(&err); err != nil {
		return err
	}
	return err
}

func GetResponse(res *http.Response, v interface{}) error {
	err := json.NewDecoder(res.Body).Decode(v)
	if err != nil {
		return err
	}

	return nil
}
