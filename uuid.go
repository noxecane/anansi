package siber

import (
	"encoding/hex"
	"fmt"

	uuid "github.com/satori/go.uuid"
)

func UUID() string {
	str := hex.EncodeToString(uuid.NewV4().Bytes())
	return fmt.Sprintf("%s-%s-%s-%s-%s", str[:8], str[8:12], str[12:16], str[16:20], str[20:])
}
