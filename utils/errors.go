package utils

import "fmt"

const (
	noProps = "NoProperties"
)

type mmError struct {
	category string
	code     string
}

func (e *mmError) Error() string {
	return fmt.Sprintf("%s: %s", e.category, e.code)
}
