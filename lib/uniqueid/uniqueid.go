package uniqueid

import (
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
)

func New() (string, error) {
	myUUID := uuid.NewV4()
	t := time.Now().Unix()
	myID := myUUID.String()
	return fmt.Sprintf("%d-%s", t, myID), nil
}
