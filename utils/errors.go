package utils

import "fmt"

const (
	NoProps  = "NoProperties"
	NoConfig = "NoConfiguration"
)

type MMError struct {
	Category string
	Code     string
}

func (e *MMError) Error() string {
	return fmt.Sprintf("%s: %s", e.Category, e.Code)
}
