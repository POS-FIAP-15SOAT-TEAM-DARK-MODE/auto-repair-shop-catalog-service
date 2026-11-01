package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusConversions(t *testing.T) {
	assert.Equal(t, "ACTIVE", ACTIVE.String())
	assert.Equal(t, "INACTIVE", INACTIVE.String())
	assert.True(t, BoolToWorkStatus(true).Bool())
	assert.False(t, BoolToWorkStatus(false).Bool())
	assert.Equal(t, ACTIVE, StringToWorkStatus("active"))
	assert.Equal(t, INACTIVE, StringToWorkStatus("whatever"))
}

func TestValidWorkStatusStringValue(t *testing.T) {
	assert.NoError(t, ValidWorkStatusStringValue("ACTIVE"))
	assert.NoError(t, ValidWorkStatusStringValue("inactive"))
	assert.ErrorIs(t, ValidWorkStatusStringValue("UNKNOWN"), ErrInvalidWorkStatusValue)
}

func TestNewWork(t *testing.T) {
	w, err := NewWork("Troca de óleo", "Troca de óleo do motor", "150.00", ACTIVE)
	require.NoError(t, err)
	assert.NotEmpty(t, w.ID)
	assert.Equal(t, "150", w.Price.String())

	_, err = NewWork("x", "y", "not-a-number", ACTIVE)
	assert.Error(t, err)
}
