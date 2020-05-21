package fakes

import (
	"fmt"
	"math/rand"

	"github.com/bxcodec/faker/v3"
	"github.com/tsaron/anansi/auth"
)

var (
	companySuffixes = [...]string{
		"Inc",
		"and Sons",
		"LLC",
		"Group",
	}
	emailProviders = [...]string{
		"gmail.com",
		"yahoo.com",
		"hotmail.com",
	}
)

func init() {
	faker.SetGenerateUniqueValues(true)
}

// Phone returns a fake nigerian phone number(without +234)
func Phone() string {
	network1 := []string{"7", "8", "9"}
	network2 := []string{"0", "1"}
	return auth.RandomFormat("0%s%s", network1, network2) + auth.RandomDigits(8)
}

// CompanyName returns a fake company name
func CompanyName() string {
	choice := rand.Intn(3)

	switch choice {
	case 1:
		return fmt.Sprintf("%s - %s", faker.LastName(), faker.LastName())
	case 2:
		return fmt.Sprintf("%s, %s and %s", faker.LastName(), faker.LastName(), faker.LastName())
	default:
		suffix := auth.RandomStringChoice(companySuffixes[:])
		return fmt.Sprintf("%s %s", faker.LastName(), suffix)
	}
}

// Email returns a fake email that ozzo can accept
func Email() string {
	// TODO: report that ozzo doesn't accept faker's emails
	provider := auth.RandomStringChoice(emailProviders[:])
	return fmt.Sprintf("%s.%s@%s", faker.FirstName(), faker.LastName(), provider)
}
