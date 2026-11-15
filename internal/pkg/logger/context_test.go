package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewConfig_CarriesServiceName(t *testing.T) {
	assert.Equal(t, "catalog-service", newConfig().InitialFields["service"])

	t.Setenv("SERVICE_NAME", "custom-name")
	assert.Equal(t, "custom-name", newConfig().InitialFields["service"])
}

func TestWith_AddsFieldsToContextLogger(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	ctx := withContext(context.Background(), &customLogger{zap.New(core)})

	ctx = With(ctx, zap.String("saga_id", "saga-1"))
	Of(ctx).Info("reservation.released")

	require.Len(t, logs.All(), 1)
	entry := logs.All()[0]
	assert.Equal(t, "reservation.released", entry.Message)
	assert.Equal(t, "saga-1", entry.ContextMap()["saga_id"])
}

func TestWith_RedactsSensitiveFields(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	ctx := withContext(context.Background(), &customLogger{zap.New(core)})

	ctx = With(ctx, zap.String("jwt", "eyJhbGciOi.secret.payload"), zap.String("Authorization", "Bearer x"))
	Of(ctx).Info("message.processed")

	require.Len(t, logs.All(), 1)
	ctxMap := logs.All()[0].ContextMap()
	assert.Equal(t, "[REDACTED]", ctxMap["jwt"])
	assert.Equal(t, "[REDACTED]", ctxMap["Authorization"])
}

func TestWith_WithoutLoggerInContextUsesGlobal(t *testing.T) {
	ctx := With(context.Background(), zap.String("message_id", "m-1"))
	assert.NotNil(t, Of(ctx))
}

func TestSanitizeFields_RedactsInternalTokenKeys(t *testing.T) {
	for _, key := range []string{"x-internal-token", "internal_token", "jwt_secret", "access_token", "jwt"} {
		assert.True(t, isSensitiveKey(key), key)
	}
	assert.False(t, isSensitiveKey("saga_id"))
}
