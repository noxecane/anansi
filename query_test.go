package siber

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"syreclabs.com/go/faker"
)

func TestParseQuery(t *testing.T) {
	type mapStruct struct {
		Name      string     `key:"name"`
		Age       int32      `key:"age" default:"30"`
		Limit     *int       `key:"limit"`
		CreatedAt *time.Time `key:"date"`
	}

	t.Run("parses a simple map to a struct", func(t *testing.T) {
		raw := map[string]string{"name": faker.Name().FirstName(), "age": faker.Number().Number(2)}
		var sample mapStruct
		if err := ParseQuery(raw, &sample); err != nil {
			t.Fatal(err)
		}

		if raw["name"] != sample.Name {
			t.Errorf("Expected name to be %s, got %s", raw["name"], sample.Name)
		}

		i, _ := strconv.Atoi(raw["age"])
		if int32(i) != sample.Age {
			t.Errorf("Expected age to be %s, got %d", raw["age"], sample.Age)
		}
	})

	t.Run("parses pointer value", func(t *testing.T) {
		raw := map[string]string{"date": "2020-09-01"}
		var sample mapStruct
		if err := ParseQuery(raw, &sample); err != nil {
			t.Fatal(err)
		}

		if sample.CreatedAt == nil {
			t.Fatal("Expected created_at to be set to a value, got nil")
		}

		if sample.CreatedAt.IsZero() {
			t.Error("Expected created_at to be set to a value, got zero value")
		}
	})

	t.Run("sets empty pointer value to nil", func(t *testing.T) {
		raw := map[string]string{}
		var sample mapStruct
		if err := ParseQuery(raw, &sample); err != nil {
			t.Fatal(err)
		}

		if sample.CreatedAt != nil {
			t.Errorf("Expected created_at to be nil, got %v", sample.CreatedAt)
		}
	})

	t.Run("zero value is not nil pointer", func(t *testing.T) {
		raw := map[string]string{"limit": "0"}
		var sample mapStruct
		if err := ParseQuery(raw, &sample); err != nil {
			t.Fatal(err)
		}

		if *sample.Limit != 0 {
			t.Errorf("Expected age to be zero value, got %d", sample.Limit)
		}
	})
}

func TestReadQuery(t *testing.T) {
	type myQuery struct {
		Account string     `key:"nuban"`
		Start   *time.Time `key:"from"`
		End     *time.Time `key:"to"`
	}

	req, err := http.NewRequest("GET", "https://sample.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	nuban := faker.Number().Number(10)
	req.URL.RawQuery = fmt.Sprintf("nuban=%s", nuban)

	var sample myQuery
	ReadQuery(req, &sample)

	if sample.Account != nuban {
		t.Errorf("Expected nuban to be %s, got %s", nuban, sample.Account)
	}

	if sample.Start != nil || sample.End != nil {
		t.Error("Expected from and to nil")
	}
}
