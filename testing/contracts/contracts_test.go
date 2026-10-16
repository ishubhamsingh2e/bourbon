package contracts

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ishubhamsingh2e/bourbon/auth"
	"github.com/ishubhamsingh2e/bourbon/cache/memcache"
	"github.com/ishubhamsingh2e/bourbon/contrib/jwt"
	"github.com/ishubhamsingh2e/bourbon/contrib/redis"
	"github.com/ishubhamsingh2e/bourbon/contrib/s3"
	"github.com/ishubhamsingh2e/bourbon/database/memdb"
	eventinprocess "github.com/ishubhamsingh2e/bourbon/events/inprocess"
	jobinprocess "github.com/ishubhamsingh2e/bourbon/jobs/inprocess"
	"github.com/ishubhamsingh2e/bourbon/mail/devmail"
	realtimeinprocess "github.com/ishubhamsingh2e/bourbon/realtime/inprocess"
	"github.com/ishubhamsingh2e/bourbon/session/memstore"
	"github.com/ishubhamsingh2e/bourbon/storage/localfs"
	"github.com/ishubhamsingh2e/bourbon/template"
	"github.com/ishubhamsingh2e/bourbon/validation"
)

func TestMemDBContractCompliance(t *testing.T) {
	db := memdb.New()
	VerifyDatabaseContract(t, db)
}

func TestMemCacheContractCompliance(t *testing.T) {
	c := memcache.New(0)
	defer c.Close()
	VerifyCacheContract(t, c)
}

func TestMemStoreContractCompliance(t *testing.T) {
	s := memstore.New(0)
	defer s.Close()
	VerifySessionStoreContract(t, s)
}

func TestHMACTokenProviderContractCompliance(t *testing.T) {
	provider := auth.NewHMACTokenProvider([]byte("contract_signing_key_456"))
	VerifyTokenProviderContract(t, provider)
}

func TestTokenGuardContractCompliance(t *testing.T) {
	provider := auth.NewHMACTokenProvider([]byte("contract_signing_key_456"))
	guard := auth.NewTokenGuard(provider)

	p := auth.NewPrincipal("contract_user").WithRoles("user")
	validToken, _ := provider.Generate(p, time.Hour)

	validReq := httptest.NewRequest(http.MethodGet, "/api", nil)
	validReq.Header.Set("Authorization", "Bearer "+validToken)

	invalidReq := httptest.NewRequest(http.MethodGet, "/api", nil)
	invalidReq.Header.Set("Authorization", "Bearer invalid.token.format")

	VerifyGuardContract(t, guard, validReq, invalidReq)
}

func TestSchemaValidatorContractCompliance(t *testing.T) {
	validator := validation.New().
		Field("email", validation.Required(), validation.Email()).
		Field("age", validation.Required(), validation.Min(18))

	validMap := map[string]any{"email": "test@example.com", "age": 21}
	invalidMap := map[string]any{"email": "bad-email", "age": 12}

	VerifyValidatorContract(t, validator, validMap, invalidMap)
}

func TestTemplateEngineContractCompliance(t *testing.T) {
	mockFS := fstest.MapFS{
		"welcome.html": &fstest.MapFile{
			Data: []byte(`<h1>Welcome {{.User}}</h1>`),
		},
	}
	engine := template.New(mockFS, "*.html", false)
	VerifyViewRendererContract(t, engine, "welcome.html", map[string]any{"User": "Gopher"}, "<h1>Welcome Gopher</h1>")
}

func TestInProcessBusContractCompliance(t *testing.T) {
	bus := eventinprocess.New(50)
	defer bus.Close()
	VerifyEventBusContract(t, bus)
}

func TestInProcessQueueContractCompliance(t *testing.T) {
	q := jobinprocess.New()
	VerifyJobQueueContract(t, q)
}

func TestInProcessHubContractCompliance(t *testing.T) {
	hub := realtimeinprocess.New()
	VerifyRealtimeHubContract(t, hub)
}

func TestLocalFSContractCompliance(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "contract_storage_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := localfs.New(tempDir)
	if err != nil {
		t.Fatalf("localfs.New failed: %v", err)
	}
	VerifyStorageContract(t, store)
}

func TestDevMailContractCompliance(t *testing.T) {
	mailer := devmail.New()
	VerifyMailerContract(t, mailer)
}

func TestRedisCacheContractCompliance(t *testing.T) {
	client := redis.NewMemoryClient()
	cacheAdapter := redis.NewCache(client)
	VerifyCacheContract(t, cacheAdapter)
}

func TestRedisSessionStoreContractCompliance(t *testing.T) {
	client := redis.NewMemoryClient()
	sessionAdapter := redis.NewSessionStore(client)
	VerifySessionStoreContract(t, sessionAdapter)
}

func TestRedisJobQueueContractCompliance(t *testing.T) {
	client := redis.NewMemoryClient()
	queueAdapter := redis.NewQueue(client)
	VerifyJobQueueContract(t, queueAdapter)
}

func TestRedisRealtimeHubContractCompliance(t *testing.T) {
	client := redis.NewMemoryClient()
	hubAdapter := redis.NewHub(client)
	VerifyRealtimeHubContract(t, hubAdapter)
}

func TestJWTTokenProviderContractCompliance(t *testing.T) {
	provider := jwt.NewTokenProvider([]byte("jwt_contract_key_1234567890123456"))
	VerifyTokenProviderContract(t, provider)
}

func TestS3StorageContractCompliance(t *testing.T) {
	s3Store := s3.New(s3.Config{Bucket: "test-bucket", Region: "us-east-1"})
	VerifyStorageContract(t, s3Store)
}
