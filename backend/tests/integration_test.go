package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"moneyvault/internal/config"
	"moneyvault/internal/encryption"
	"moneyvault/internal/handlers"
	"moneyvault/internal/middleware"
	"moneyvault/internal/repositories"
	"moneyvault/internal/services"
	"moneyvault/tests/testhelpers"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRouter(db *sqlx.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		JWTSecret:          "test-jwt-secret-at-least-32-chars!",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Environment:        "test",
	}

	enc := encryption.NewService()

	userRepo := repositories.NewUserRepository(db)
	catRepo := repositories.NewCategoryRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)
	dekRepo := repositories.NewDEKSessionRepository(db)
	acctRepo := repositories.NewAccountRepository(db)
	txRepo := repositories.NewTransactionRepository(db)

	authService := services.NewAuthService(userRepo, catRepo, tokenRepo, dekRepo, enc, cfg)
	acctService := services.NewAccountService(acctRepo, enc)
	txService := services.NewTransactionService(txRepo, acctRepo, enc)

	authHandler := handlers.NewAuthHandler(authService)
	acctHandler := handlers.NewAccountHandler(acctService)
	txHandler := handlers.NewTransactionHandler(txService)

	r := gin.New()
	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	protected := api.Group("")
	protected.Use(middleware.Auth(authService))

	protected.GET("/auth/me", authHandler.Me)
	protected.POST("/accounts", acctHandler.Create)
	protected.GET("/accounts", acctHandler.List)
	protected.POST("/transactions", txHandler.Create)
	protected.GET("/transactions", txHandler.List)

	return r
}

func TestIntegration_FullFlow(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	testhelpers.TruncateTables(t, db)

	router := setupRouter(db)

	// 1. Register
	registerBody, _ := json.Marshal(map[string]string{
		"email":    "fullflow@test.com",
		"password": "SecurePass1!",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "register should return 201: %s", w.Body.String())

	// 2. Login
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "fullflow@test.com",
		"password": "SecurePass1!",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "login should return 200: %s", w.Body.String())

	var loginResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	accessToken, ok := loginResp["access_token"].(string)
	require.True(t, ok, "response should contain access_token")
	require.NotEmpty(t, accessToken)

	// 3. Get current user
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var meResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &meResp)
	userMap := meResp["user"].(map[string]interface{})
	assert.Equal(t, "fullflow@test.com", userMap["email"])

	// 4. Create Account
	acctBody, _ := json.Marshal(map[string]interface{}{
		"name":     "Test Checking",
		"type":     "checking",
		"currency": "USD",
		"balance":  "5000.00",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/accounts", bytes.NewReader(acctBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "create account should return 201: %s", w.Body.String())

	var acctResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &acctResp)
	acctMap := acctResp["account"].(map[string]interface{})
	accountID := acctMap["id"].(string)
	assert.Equal(t, "Test Checking", acctMap["name"])

	// 5. List Accounts
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var listAcctResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listAcctResp)
	accounts := listAcctResp["accounts"].([]interface{})
	assert.Len(t, accounts, 1)

	// 6. Create Transaction
	txBody, _ := json.Marshal(map[string]interface{}{
		"account_id":  accountID,
		"type":        "expense",
		"amount":      "42.50",
		"currency":    "USD",
		"description": "Integration test purchase",
		"date":        "2025-06-15",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/transactions", bytes.NewReader(txBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "create transaction should return 201: %s", w.Body.String())

	var txResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &txResp)
	txMap := txResp["transaction"].(map[string]interface{})
	assert.Equal(t, "42.50", txMap["amount"])
	assert.Equal(t, "Integration test purchase", txMap["description"])

	// 7. List Transactions
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var listTxResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listTxResp)
	txData := listTxResp["data"].([]interface{})
	assert.Len(t, txData, 1)

	// 8. Verify unauthenticated access is rejected
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/accounts", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
