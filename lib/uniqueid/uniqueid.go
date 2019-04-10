package uniqueid

import (
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
)

func New() (string, error) {
	myUUID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	t := time.Now().Unix()
	myID := myUUID.String()
	return fmt.Sprintf("%d-%s", t, myID), nil
}
