package provider

import (
	"fmt"
	"strings"
)

func errNotFound(what string) error {
	return fmt.Errorf("%s not returned by the Unraid API (missing permission or not available)", what)
}

// withPermissionHint explains the API's terse "Forbidden resource" error.
func withPermissionHint(err error) string {
	msg := err.Error()
	if strings.Contains(msg, "Forbidden") {
		msg += "\n\nThe API key lacks the permission this data source needs (see \"Requires ...\" in its documentation). " +
			"Grant that resource/action to the key or use a key with a broader role."
	}
	return msg
}
