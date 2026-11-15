//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app/bootstrap"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/messaging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const internalToken = "integration-internal-token"

// env returns the variable or the local default used by docker-compose / CI.
func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// testEnv is a running catalog service backed by real Postgres and (optionally) SQS.
type testEnv struct {
	db     *sql.DB
	server *httptest.Server
	sqs    *sqs.Client
	queue  string
}

// setup starts the HTTP API and the ReleaseStock consumer against real dependencies.
// Migrations must already be applied to the target database.
func setup(t *testing.T) *testEnv {
	t.Helper()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		env("POSTGRES_USER", "postgres"), env("POSTGRES_PASSWORD", "postgres"),
		env("POSTGRES_HOST", "localhost"), env("POSTGRES_PORT", "5432"),
		env("POSTGRES_DB", "catalog"), env("POSTGRES_SSLMODE", "disable"))
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() { _ = db.Close() })
	app.ConnectWithDB(db)

	t.Setenv("INTERNAL_TOKEN", internalToken)
	router := bootstrap.NewRouter(db, internalToken, &container.Middlewares{})
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	te := &testEnv{db: db, server: server}
	te.startConsumer(t)
	return te
}

// startConsumer creates a dedicated queue and runs the real consumer wiring from bootstrap.
func (te *testEnv) startConsumer(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	client, err := messaging.NewSQSClient(ctx)
	require.NoError(t, err)
	te.sqs = client

	name := "catalog-it-" + uuid.NewString()
	out, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: aws.String(name)})
	require.NoError(t, err)
	te.queue = aws.ToString(out.QueueUrl)
	t.Setenv("CATALOG_QUEUE_URL", te.queue)

	server, err := bootstrap.Consumer(ctx)
	require.NoError(t, err)
	require.NotNil(t, server)

	done := make(chan error, 1)
	go func() { done <- server.Start(ctx) }()
	t.Cleanup(func() {
		stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_ = server.Stop(stopCtx)
		select {
		case <-done:
		case <-time.After(30 * time.Second):
		}
		_, _ = client.DeleteQueue(ctx, &sqs.DeleteQueueInput{QueueUrl: aws.String(te.queue)})
	})
}

// send publishes a raw message body to the consumer queue.
func (te *testEnv) send(t *testing.T, body string) {
	t.Helper()
	_, err := te.sqs.SendMessage(context.Background(), &sqs.SendMessageInput{QueueUrl: aws.String(te.queue), MessageBody: aws.String(body)})
	require.NoError(t, err)
}

// token returns a development JWT for the given role.
func token(t *testing.T, role auth.Role) string {
	t.Helper()
	tok, err := auth.GenerateToken(uuid.NewString(), []auth.Role{role}, time.Now().Add(time.Hour))
	require.NoError(t, err)
	return tok
}

// call performs a request against the running API and returns the status code and body.
func (te *testEnv) call(t *testing.T, method, path, body string, headers map[string]string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(method, te.server.URL+path, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(b)
}

// insertSupply creates a supply with the given stock and returns its id.
func (te *testEnv) insertSupply(t *testing.T, stock int) string {
	t.Helper()
	id := uuid.NewString()
	_, err := te.db.Exec(`INSERT INTO supply (id, name, description, unit_price, stock_quantity) VALUES ($1, $2, $3, $4, $5)`,
		id, "Filtro "+id[:8], "Filtro de teste de integração", "35.00", stock)
	require.NoError(t, err)
	return id
}

func (te *testEnv) stock(t *testing.T, supplyID string) int {
	t.Helper()
	var n int
	require.NoError(t, te.db.QueryRow(`SELECT stock_quantity FROM supply WHERE id = $1`, supplyID).Scan(&n))
	return n
}
