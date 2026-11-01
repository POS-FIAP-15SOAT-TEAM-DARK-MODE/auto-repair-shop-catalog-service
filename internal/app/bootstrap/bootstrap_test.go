package bootstrap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter_ServesPingAndProtectsInternal(t *testing.T) {
	router := NewRouter(nil, "token", middlewaresContainer())

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	assert.Equal(t, http.StatusOK, rec.Code)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/internal/works/1", nil))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestConsumer_DisabledWithoutQueueURL(t *testing.T) {
	t.Setenv("CATALOG_QUEUE_URL", "")

	server, err := Consumer(context.Background())

	require.NoError(t, err)
	assert.Nil(t, server)
}
