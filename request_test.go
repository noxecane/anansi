package anansi

import (
	"net/http"
	"strconv"
	"testing"

	ozzo "github.com/go-ozzo/ozzo-validation/v4"
	"syreclabs.com/go/faker"
)

type userQuery struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (u *userQuery) Validate() error {
	return ozzo.ValidateStruct(u,
		ozzo.Field(&u.Name, ozzo.Required),
		ozzo.Field(&u.Age, ozzo.Required),
	)
}

func TestReadQuery(t *testing.T) {
	fName := faker.Name().FirstName()
	ageOne := faker.Number().Number(2)
	ageTwo := faker.Number().Number(2)

	req, err := http.NewRequest("GET", "https://google.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	q := req.URL.Query()

	q.Add("name", fName)
	q.Add("age", ageOne)
	q.Add("age", ageTwo)

	req.URL.RawQuery = q.Encode()

	uQ := new(userQuery)
	ReadQuery(req, uQ)

	if uQ.Name != fName {
		t.Errorf("Expected name %s. got name %s", fName, uQ.Name)
	}

	uQAge, err := strconv.Atoi(ageOne)
	if err != nil {
		t.Fatal(err)
	}

	if uQ.Age != uQAge {
		t.Errorf("Expected name %s. got name %s", fName, uQ.Name)
	}
}
