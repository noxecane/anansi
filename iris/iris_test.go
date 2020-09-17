package iris

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/segmentio/encoding/json"
	"syreclabs.com/go/faker"
)

type jSMock struct {
	Name    string   `json:"name"`
	Company string   `json:"company"`
	Emails  []string `json:"emails"`
}

func TestGetResponse(t *testing.T) {
	name := faker.Name().FirstName()

	t.Run("It should decode response", func(t *testing.T) {
		b, err := json.Marshal(map[string]interface{}{
			"data": map[string]interface{}{"name": name},
		})
		if err != nil {
			panic(err)
		}

		res := http.Response{
			Body: ioutil.NopCloser(bytes.NewBuffer(b)),
		}

		var data jSMock
		err = GetResponse(&res, &data)
		if err != nil {
			t.Fatal(errors.Wrap(err, "error getting response"))
		}

		if data.Name != name {
			t.Fatalf("expected %s got %s", name, data.Name)
		}
	})
}
