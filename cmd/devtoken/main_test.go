package main

import (
	"bytes"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_PrintsValidToken(t *testing.T) {
	var out bytes.Buffer
	require.NoError(t, run([]string{"-role", "MECHANIC"}, &out))

	claims, err := auth.GetClaims(trimNewline(out.String()))
	require.NoError(t, err)
	assert.Equal(t, defaultUserID, claims.UserId)
	assert.Equal(t, []auth.Role{auth.MECHANIC}, claims.Roles)
}

func TestRun_RejectsUnknownRole(t *testing.T) {
	var out bytes.Buffer
	assert.ErrorIs(t, run([]string{"-role", "ROOT"}, &out), errUnknownRole)
	assert.Empty(t, out.String())
}

func TestRun_InvalidFlag(t *testing.T) {
	var out bytes.Buffer
	assert.Error(t, run([]string{"-nope"}, &out))
}

func trimNewline(s string) string {
	return s[:len(s)-1]
}
