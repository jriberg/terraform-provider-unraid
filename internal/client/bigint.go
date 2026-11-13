package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// BigInt decodes the GraphQL BigInt scalar, which may arrive as a JSON number or a numeric string.
type BigInt int64

func (b *BigInt) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, `"`)
	if string(data) == "null" {
		return nil
	}
	n, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid BigInt %q: %w", data, err)
	}
	*b = BigInt(n)
	return nil
}

func (b BigInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(b))
}
