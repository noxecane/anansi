package ajax

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/tsaron/anansi/jwt"
	"syreclabs.com/go/faker"
)

type mockStruct struct {
	Name    string   `json:"name"`
	Company string   `json:"company"`
	Emails  []string `json:"emails"`
}

func TestBearerToken(t *testing.T) {
	tokenValue := faker.Lorem().Characters(32)

	req, err := http.NewRequest("GET", "some-url", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenValue)
	req.Header.Set("X-Request-ID", faker.Lorem().Characters(16))
	fmt.Println(req.Header)

	client := NewClient(Config{
		Secret:         []byte(faker.Lorem().Characters(32)),
		Service:        "some-service",
		HeadlessScheme: "some-scheme",
	})

	token, err := client.BearerToken(req)
	if err != nil {
		t.Fatal(err)
	}

	ajaxReq, err := client.NewRequest(req, "GET", "some-other-url", token, nil)
	if err != nil {
		t.Fatal(err)
	}

	if ajaxReq.Header.Get("Authorization") != "Bearer "+tokenValue {
		t.Errorf("Expected authorization to be \"Bearer %s\", got %s", tokenValue, ajaxReq.Header.Get("Authorization"))
	}
}

func TestHeadlessToken(t *testing.T) {
	secret := []byte(faker.Lorem().Characters(32))
	client := NewClient(Config{
		Secret:         secret,
		Service:        "some-service",
		HeadlessScheme: "some-scheme",
	})

	type sessionInfo struct {
		Info string
	}
	session := sessionInfo{"some-session-info"}

	token, err := client.HeadlessToken(session)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "some-url", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Request-ID", faker.Lorem().Characters(16))

	ajaxReq, err := client.NewRequest(req, "GET", "some-other-url", token, nil)
	if err != nil {
		t.Fatal(err)
	}

	if ajaxReq.Header.Get("Authorization") != fmt.Sprintf("%s %s", client.headlessScheme, token.value) {
		t.Fatalf("Expected authorization to be \"%s %s\", got %s", client.headlessScheme, token.value, ajaxReq.Header.Get("Authorization"))
	}

	auth := strings.Split(ajaxReq.Header.Get("Authorization"), " ")
	tokenStr := strings.TrimSpace(auth[1])

	var loadedSession sessionInfo
	if err := jwt.DecodeStruct(secret, []byte(tokenStr), &loadedSession); err != nil {
		t.Fatal(err)
	}

	if loadedSession.Info != session.Info {
		t.Errorf("Expected session to store %s, got %s", session.Info, loadedSession.Info)
	}
}

func TestGetResponse(t *testing.T) {
	name := faker.Name().FirstName()
	b, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		panic(err)
	}

	res := http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer(b)),
	}

	var data mockStruct
	err = GetResponse(&res, &data)
	if err != nil {
		t.Fatal(errors.Wrap(err, "error getting response"))
	}

	if data.Name != name {
		t.Fatalf("expected %s got %s", name, data.Name)
	}
}
