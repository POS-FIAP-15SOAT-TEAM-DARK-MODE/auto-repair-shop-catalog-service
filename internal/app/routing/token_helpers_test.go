package routing

import (
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/auth"
)

func timeNowPlusHour() time.Time { return time.Now().Add(time.Hour) }

func tokenFor(t *testing.T, role auth.Role) string {
	t.Helper()
	token, err := auth.GenerateToken("u", []auth.Role{role}, timeNowPlusHour())
	if err != nil {
		t.Fatal(err)
	}
	return token
}
