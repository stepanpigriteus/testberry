package handlemessage

import (
	"context"
	"encoding/json"
)

func HandleOrderMessage(ctx context.Context, msg []byte) error {
	var order map[string]interface{}
	if err := json.Unmarshal(msg, &order); err != nil {
		return err
	}

	return nil
}
