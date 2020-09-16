package siber

import "time"

var ISO_FORMATS = [3]string{"2006-01-02T15:04:05.000Z", "2006-01-02", "2006-01-02T15:04:05"}

func FromISO(date string) (time.Time, error) {
	var err error

	for _, format := range ISO_FORMATS {
		createdAt, err := time.Parse(format, date)
		if err == nil {
			return createdAt, nil
		}
	}

	return time.Time{}, err
}
