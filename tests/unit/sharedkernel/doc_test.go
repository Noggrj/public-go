package sharedkernel_test

import (
	"testing"

	"github.com/noggrj/autorepair/internal/sharedkernel"
)

func TestNewDocumentoBR(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid CPF", "12345678909", false},
		{"Valid CNPJ", "12345678000195", false},
		{"Invalid Length", "123", true},
		{"With Formatting", "123.456.789-09", false}, // Regex strips non-digits
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sharedkernel.NewDocumentoBR(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDocumentoBR() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
