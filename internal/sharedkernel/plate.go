package sharedkernel

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidPlate = errors.New("invalid plate format")
)

type PlacaBR struct {
	value string
}

func NewPlacaBR(plate string) (PlacaBR, error) {
	upper := strings.ToUpper(strings.ReplaceAll(plate, "-", ""))
	
	// Legacy: AAA1234 (3 letters, 4 numbers)
	// Mercosul: AAA1A23 (3 letters, 1 number, 1 letter, 2 numbers)
	legacyRegex := regexp.MustCompile(`^[A-Z]{3}[0-9]{4}$`)
	mercosulRegex := regexp.MustCompile(`^[A-Z]{3}[0-9][A-Z][0-9]{2}$`)

	if !legacyRegex.MatchString(upper) && !mercosulRegex.MatchString(upper) {
		return PlacaBR{}, ErrInvalidPlate
	}

	return PlacaBR{value: upper}, nil
}

func (p PlacaBR) String() string {
	return p.value
}

func (p PlacaBR) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.value)
}
