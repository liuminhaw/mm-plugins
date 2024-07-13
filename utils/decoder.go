package utils

import (
	"fmt"
	"net/url"

	"github.com/liuminhaw/mist-miner/shared"
)

// DocumentUrlDecode decodes a json formatted URL-encoded string and returns the decoded result.
func DocumentUrlDecode(content string) (string, error) {
	decodeContent, err := url.QueryUnescape(content)
	if err != nil {
		return "", fmt.Errorf("DocumentUrlDecode: %w", err)
	}

	cleanContent, err := shared.JsonNormalize(decodeContent)
	if err != nil {
		return "", fmt.Errorf("DocumentUrlDecode: %w", err)
	}

	return string(cleanContent), nil
}
