package sharedkernel

import (
	"encoding/json"
	"errors"
	"regexp"
)

var (
	ErrInvalidDocument = errors.New("invalid document format")
)

type DocumentoBR struct {
	value string
}

func NewDocumentoBR(doc string) (DocumentoBR, error) {
	clean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(doc, "")
	
	if len(clean) == 11 {
		if !isValidCPF(clean) {
			return DocumentoBR{}, ErrInvalidDocument
		}
	} else if len(clean) == 14 {
		if !isValidCNPJ(clean) {
			return DocumentoBR{}, ErrInvalidDocument
		}
	} else {
		return DocumentoBR{}, ErrInvalidDocument
	}
	
	return DocumentoBR{value: clean}, nil
}

func isValidCPF(cpf string) bool {
	// Check for known invalid patterns (all digits equal)
	if isAllDigitsEqual(cpf) {
		return false
	}

	// Calculate first verifier digit
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(cpf[i]-'0') * (10 - i)
	}
	remainder := sum % 11
	digit1 := 0
	if remainder >= 2 {
		digit1 = 11 - remainder
	}

	if int(cpf[9]-'0') != digit1 {
		return false
	}

	// Calculate second verifier digit
	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(cpf[i]-'0') * (11 - i)
	}
	remainder = sum % 11
	digit2 := 0
	if remainder >= 2 {
		digit2 = 11 - remainder
	}

	return int(cpf[10]-'0') == digit2
}

func isValidCNPJ(cnpj string) bool {
	// Check for known invalid patterns
	if isAllDigitsEqual(cnpj) {
		return false
	}

	// First digit validation
	weights1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 12; i++ {
		sum += int(cnpj[i]-'0') * weights1[i]
	}
	remainder := sum % 11
	digit1 := 0
	if remainder >= 2 {
		digit1 = 11 - remainder
	}

	if int(cnpj[12]-'0') != digit1 {
		return false
	}

	// Second digit validation
	weights2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum = 0
	for i := 0; i < 13; i++ {
		sum += int(cnpj[i]-'0') * weights2[i]
	}
	remainder = sum % 11
	digit2 := 0
	if remainder >= 2 {
		digit2 = 11 - remainder
	}

	return int(cnpj[13]-'0') == digit2
}

func isAllDigitsEqual(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

func (d DocumentoBR) String() string {
	return d.value
}

func (d DocumentoBR) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.value)
}
