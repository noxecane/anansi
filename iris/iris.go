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

type jSendSuccess struct {
	Data map[string]interface{} `json:"data"`
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
	token, err := jwt.EncodeEmbedded(c.serviceSecret, c.headlessDuration, v)
	if err != nil {
		return Token{}, err
	}

	return Token{token, true}, nil
}

// NewRequest is a wrapper around http.NNewRequest that adds the required
// headers for distributed tracing. The requests will only last as long as the parent
// request(it uses the request's context)
func (c *Client) NewRequest(r *http.Request, method, url string, token Token, body io.Reader) (*http.Request, error) {
	reqId := r.Header.Get("X-Request-Id")
	if reqId == "" {
		return nil, errors.New("request ID not set on base request")
	}

	if token.value == "" {
		return nil, errors.New("authentication token is not set")
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

// NewRequestWitchContext is the same as NewRequest, only the user can now control
// how long before the request times out.
func (c *Client) NewRequestWithContext(ctx context.Context, r *http.Request, method, url string, token Token, body io.Reader) (*http.Request, error) {
	reqId := r.Header.Get("X-Request-Id")
	if reqId == "" {
		return nil, errors.New("request ID not set on base request")
	}

	if token.value == "" {
		return nil, errors.New("authentication token is not set")
	}

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
