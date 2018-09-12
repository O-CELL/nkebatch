package nkebatch

import "fmt"

// Error Management

/* Type defintion */
const (
	ERRUNDEF        = 0
	ERRNOTSUPPORTED = iota
	ERROUTOFRANGE
	ERRNOTSUPPORTEDCTS
	ERRHEADER
	ERRINVALIDCONFIG
)

var mapErrorMessage = map[uint]string{
	ERRUNDEF:           "Undefined",
	ERRNOTSUPPORTED:    "Not supported",
	ERROUTOFRANGE:      "Out Of RANGE",
	ERRNOTSUPPORTEDCTS: "CTS not supported",
	ERRHEADER:          "Wrong Header",
	ERRINVALIDCONFIG:	"Invalid configuration",
}

// Error ...
type Error struct {
	id     int
	reason string
}

// Error...
func (e *Error) Error() string {
	return fmt.Sprintf("%d - %s", e.id, e.reason)
}
