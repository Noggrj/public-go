package sharedkernel_test

import (
	"encoding/json"
	"testing"

	"github.com/noggrj/autorepair/internal/sharedkernel"
	"github.com/stretchr/testify/assert"
)

func TestNewPlacaBR(t *testing.T) {
	// Valid Old Format
	p, err := sharedkernel.NewPlacaBR("ABC-1234")
	assert.NoError(t, err)
	assert.Equal(t, "ABC1234", p.String())

	// Valid Mercosul Format
	p, err = sharedkernel.NewPlacaBR("ABC1D23")
	assert.NoError(t, err)
	assert.Equal(t, "ABC1D23", p.String())

	// Invalid
	_, err = sharedkernel.NewPlacaBR("INVALID")
	assert.Error(t, err)

	// JSON Marshalling
	b, err := json.Marshal(p)
	assert.NoError(t, err)
	assert.Equal(t, `"ABC1D23"`, string(b))
}

func TestDocumentoBR_JSON(t *testing.T) {
	d, _ := sharedkernel.NewDocumentoBR("12345678909")
	b, err := json.Marshal(d)
	assert.NoError(t, err)
	assert.Equal(t, `"12345678909"`, string(b))
	assert.Equal(t, "12345678909", d.String())
}
