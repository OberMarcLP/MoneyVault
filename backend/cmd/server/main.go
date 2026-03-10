package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"moneyvault/internal/config"
	"moneyvault/internal/encryption"
	"moneyvault/internal/handlers"
	"moneyvault/internal/integrations"
	"moneyvault/internal/middleware"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"
	"moneyvault/internal/services"

	_ "moneyvault/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title MoneyVault API
// @version 1.0
// @description Self-hosted personal finance API with encryption, investments, crypto, and analytics.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}"
func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	repositories.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath)

	db := repositories.NewDB(cfg.DatabaseURL)
	defer db.Close()

	enc := encryption.NewService()
	enc.SetSessionKey(cfg.JWTSecret)

	userRepo := repositories.NewUserRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	transactionRepo := repositories.NewTransactionRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	budgetRepo := repositories.NewBudgetRepository(db)
	recurringRepo := repositories.NewRecurringRepository(db)
	importRepo := repositories.NewImportRepository(db)
	holdingRepo := repositories.NewHoldingRepository(db)
	walletRepo := repositories.NewWalletRepository(db)
	analyticsRepo := repositories.NewAnalyticsRepository(db)
	notifRepo := repositories.NewNotificationRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)
	resetRepo := repositories.NewPasswordResetRepository(db)
	auditRepo := repositories.NewAuditRepository(db)
	dekSessionRepo := repositories.NewDEKSessionRepository(db)

	yahooClient := integrations.NewYahooClient()
	coingeckoClient := integrations.NewCoinGeckoClient()
	etherscanClient := integrations.NewEtherscanClient()
	exchangeRateClient := integrations.NewExchangeRateClient()
	binanceClient := integrations.NewBinanceClient()
	coinbaseClient := integrations.NewCoinbaseClient()
	krakenClient := integrations.NewKrakenClient()

	authService := services.NewAuthService(userRepo, categoryRepo, tokenRepo, dekSessionRepo, enc, cfg)
	accountService := services.NewAccountService(accountRepo, enc)
	transactionService := services.NewTransactionService(transactionRepo, accountRepo, enc)
	categoryService := services.NewCategoryService(categoryRepo)
	budgetService := services.NewBudgetService(budgetRepo, categoryRepo, transactionRepo, enc)
	recurringService := services.NewRecurringService(recurringRepo, transactionRepo, accountRepo, enc)
	importService := services.NewImportService(importRepo, transactionRepo, categoryRepo, enc)
	dividendRepo := repositories.NewDividendRepository(db)
	investmentService := services.NewInvestmentService(holdingRepo, dividendRepo, yahooClient)
	cryptoService := services.NewCryptoService(holdingRepo, walletRepo, coingeckoClient, etherscanClient)
	analyticsService := services.NewAnalyticsService(analyticsRepo, accountRepo, holdingRepo, categoryRepo, budgetRepo, recurringRepo, transactionRepo, enc, exchangeRateClient)
	notifService := services.NewNotificationService(notifRepo, budgetRepo, holdingRepo, analyticsRepo, categoryRepo, transactionRepo, enc)
	exportService := services.NewExportService(transactionRepo, accountRepo, categoryRepo, enc)
	passwordResetService := services.NewPasswordResetService(resetRepo, userRepo, tokenRepo, enc)
	exchangeConnRepo := repositories.NewExchangeConnectionRepository(db)
	pushRepo := repositories.NewPushRepository(db)
	webauthnRepo := repositories.NewWebAuthnRepository(db)
	exchangeConnService := services.NewExchangeConnectionService(exchangeConnRepo, enc, binanceClient, coinbaseClient, krakenClient)
	pushService := services.NewPushService(pushRepo, cfg)
	notifService.SetPushService(pushService)
	webauthnService, err := services.NewWebAuthnService(cfg, webauthnRepo, userRepo, authService)
	if err != nil {
		log.Fatalf("Failed to initialize WebAuthn: %v", err)
	}

	// Restore DEK sessions from database (survives server restarts)
	if count := authService.RestoreAllSessions(); count > 0 {
		log.Printf("Restored %d active DEK sessions", count)
	}

	// Background jobs with graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	startJob := func(name string, initialDelay, interval time.Duration, fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if initialDelay > 0 {
				select {
				case <-time.After(initialDelay):
				case <-ctx.Done():
					return
				}
			}
			fn()
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					fn()
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Fetch exchange rates immediately on startup, then every hour
	startJob("exchange-rates", 0, 1*time.Hour, func() {
		if err := exchangeRateClient.FetchRates(); err != nil {
			log.Printf("Exchange rate fetch error: %v", err)
		} else {
			log.Println("Exchange rates refreshed")
		}
	})

	startJob("recurring", 0, 1*time.Hour, func() {
		if count, err := recurringService.ProcessDue(); err == nil && count > 0 {
			log.Printf("Created %d recurring transactions", count)
		}
	})

	startJob("stock-prices", 5*time.Second, 15*time.Minute, func() {
		if err := investmentService.RefreshPrices(); err != nil {
			log.Printf("Stock price refresh error: %v", err)
		}
	})

	startJob("crypto-prices", 10*time.Second, 5*time.Minute, func() {
		if err := cryptoService.RefreshCryptoPrices(); err != nil {
			log.Printf("Crypto price refresh error: %v", err)
		}
	})

	startJob("snapshots", 30*time.Second, 24*time.Hour, func() {
		analyticsService.TakeAllSnapshots()
	})

	startJob("alerts", 1*time.Minute, 30*time.Minute, func() {
		notifService.EvaluateAlerts()
	})

	startJob("cleanup", 1*time.Hour, 1*time.Hour, func() {
		if count, err := authService.CleanupExpiredTokens(); err == nil && count > 0 {
			log.Printf("Cleaned up %d expired revoked tokens", count)
		}
		if count, err := passwordResetService.CleanupExpired(); err == nil && count > 0 {
			log.Printf("Cleaned up %d expired password reset tokens", count)
		}
		if count, err := authService.CleanupExpiredDEKSessions(); err == nil && count > 0 {
			log.Printf("Cleaned up %d expired DEK sessions", count)
		}
	})

	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(authService)
	accountHandler := handlers.NewAccountHandler(accountService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	budgetHandler := handlers.NewBudgetHandler(budgetService)
	recurringHandler := handlers.NewRecurringHandler(recurringService)
	importHandler := handlers.NewImportHandler(importService)
	adminHandler := handlers.NewAdminHandler(userRepo)
	investmentHandler := handlers.NewInvestmentHandler(investmentService)
	cryptoHandler := handlers.NewCryptoHandler(cryptoService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	notifHandler := handlers.NewNotificationHandler(notifService)
	passwordResetHandler := handlers.NewPasswordResetHandler(passwordResetService)
	exportHandler := handlers.NewExportHandler(exportService)
	webauthnHandler := handlers.NewWebAuthnHandler(webauthnService, authService)
	exchangeHandler := handlers.NewExchangeHandler(exchangeRateClient)
	exchangeConnHandler := handlers.NewExchangeConnectionHandler(exchangeConnService)
	pushHandler := handlers.NewPushHandler(pushService)
	auditHandler := handlers.NewAuditHandler(auditRepo)
	e2eService := services.NewE2EService(db, userRepo, enc)
	e2eHandler := handlers.NewE2EHandler(e2eService)

	r := gin.Default()

	if cfg.TrustedProxies != nil {
		if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
			log.Fatalf("Invalid TRUSTED_PROXIES configuration: %v", err)
		}
		log.Printf("Trusted proxies configured: %v", cfg.TrustedProxies)
	}

	r.Use(middleware.ErrorHandler())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.SecurityHeaders())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Data routes get a generous rate limit; auth routes get a stricter one
	r.Use(middleware.RateLimit(cfg.RateLimitPerMinute))

	r.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")

	api.GET("/health", healthHandler.Health)

	auth := api.Group("/auth")
	auth.Use(middleware.AuthRateLimit(20))
	{
		auth.POST("/register", middleware.Audit(auditRepo, "register", "user"), authHandler.Register)
		auth.POST("/login", middleware.Audit(auditRepo, "login", "user"), authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/password-reset/request", middleware.Audit(auditRepo, "password_reset_request", "user"), passwordResetHandler.RequestReset)
		auth.POST("/password-reset/confirm", middleware.Audit(auditRepo, "password_reset_confirm", "user"), passwordResetHandler.ConfirmReset)
		auth.POST("/webauthn/login/begin", webauthnHandler.BeginLogin)
		auth.POST("/webauthn/login/finish", webauthnHandler.FinishLogin)
	}

	protected := api.Group("")
	protected.Use(middleware.Auth(authService))
	{
		protected.POST("/auth/logout", middleware.Audit(auditRepo, "logout", "user"), authHandler.Logout)
		protected.GET("/auth/me", authHandler.Me)
		protected.PUT("/auth/preferences", authHandler.UpdatePreferences)
		protected.POST("/auth/totp/setup", authHandler.SetupTOTP)
		protected.POST("/auth/totp/verify", authHandler.VerifyTOTP)
		protected.DELETE("/auth/totp", authHandler.DisableTOTP)
		protected.POST("/auth/verify-email", authHandler.VerifyEmail)
		protected.POST("/auth/webauthn/register/begin", webauthnHandler.BeginRegistration)
		protected.POST("/auth/webauthn/register/finish", webauthnHandler.FinishRegistration)
		protected.GET("/auth/webauthn/credentials", webauthnHandler.ListCredentials)
		protected.DELETE("/auth/webauthn/credentials/:id", webauthnHandler.DeleteCredential)

		accounts := protected.Group("/accounts")
		{
			accounts.POST("", accountHandler.Create)
			accounts.GET("", accountHandler.List)
			accounts.GET("/:id", accountHandler.GetByID)
			accounts.PUT("/:id", accountHandler.Update)
			accounts.DELETE("/:id", accountHandler.Delete)
		}

		transactions := protected.Group("/transactions")
		{
			transactions.POST("", transactionHandler.Create)
			transactions.GET("", transactionHandler.List)
			transactions.GET("/:id", transactionHandler.GetByID)
			transactions.PUT("/:id", transactionHandler.Update)
			transactions.DELETE("/:id", transactionHandler.Delete)
		}

		categories := protected.Group("/categories")
		{
			categories.POST("", categoryHandler.Create)
			categories.GET("", categoryHandler.List)
			categories.GET("/:id", categoryHandler.GetByID)
			categories.PUT("/:id", categoryHandler.Update)
			categories.DELETE("/:id", categoryHandler.Delete)
		}

		budgets := protected.Group("/budgets")
		{
			budgets.POST("", budgetHandler.Create)
			budgets.GET("", budgetHandler.List)
			budgets.GET("/:id", budgetHandler.GetByID)
			budgets.PUT("/:id", budgetHandler.Update)
			budgets.DELETE("/:id", budgetHandler.Delete)
		}

		recurring := protected.Group("/recurring")
		{
			recurring.POST("", recurringHandler.Create)
			recurring.GET("", recurringHandler.List)
			recurring.GET("/:id", recurringHandler.GetByID)
			recurring.DELETE("/:id", recurringHandler.Delete)
			recurring.POST("/:id/toggle", recurringHandler.Toggle)
		}

		investments := protected.Group("/investments")
		{
			investments.POST("", investmentHandler.Create)
			investments.GET("", investmentHandler.List)
			investments.GET("/summary", investmentHandler.Summary)
			investments.GET("/gains", investmentHandler.RealizedGains)
			investments.POST("/refresh-prices", investmentHandler.RefreshPrices)
			investments.GET("/price-history/:symbol", investmentHandler.PriceHistory)
			investments.GET("/:id", investmentHandler.GetByID)
			investments.PUT("/:id", investmentHandler.Update)
			investments.DELETE("/:id", investmentHandler.Delete)
			investments.POST("/:id/sell", investmentHandler.Sell)
			investments.POST("/dividends", investmentHandler.CreateDividend)
			investments.GET("/dividends", investmentHandler.ListDividends)
			investments.GET("/dividends/summary", investmentHandler.GetDividendSummary)
			investments.DELETE("/dividends/:id", investmentHandler.DeleteDividend)
		}

		crypto := protected.Group("/crypto")
		{
			crypto.GET("/summary", cryptoHandler.Summary)
			crypto.GET("/search", cryptoHandler.SearchTokens)
			crypto.POST("/refresh-prices", cryptoHandler.RefreshCryptoPrices)

			crypto.POST("/wallets", cryptoHandler.CreateWallet)
			crypto.GET("/wallets", cryptoHandler.ListWallets)
			crypto.DELETE("/wallets/:id", cryptoHandler.DeleteWallet)
			crypto.POST("/wallets/:id/sync", cryptoHandler.SyncWallet)
			crypto.GET("/wallets/:id/transactions", cryptoHandler.WalletTransactions)
			crypto.GET("/wallet-transactions", cryptoHandler.AllWalletTransactions)
		}

		analytics := protected.Group("/analytics")
		{
			analytics.GET("/net-worth", analyticsHandler.NetWorthHistory)
			analytics.GET("/spending", analyticsHandler.SpendingBreakdown)
			analytics.GET("/trends", analyticsHandler.SpendingTrends)
			analytics.GET("/top-expenses", analyticsHandler.TopExpenses)
			analytics.GET("/budget-vs-actual", analyticsHandler.BudgetVsActual)
			analytics.GET("/cash-flow", analyticsHandler.CashFlowForecast)
			analytics.GET("/asset-allocation", analyticsHandler.AssetAllocation)
		}

		notifications := protected.Group("/notifications")
		{
			notifications.GET("", notifHandler.List)
			notifications.GET("/count", notifHandler.UnreadCount)
			notifications.POST("/:id/read", notifHandler.MarkRead)
			notifications.POST("/read-all", notifHandler.MarkAllRead)
			notifications.DELETE("/:id", notifHandler.Delete)
			notifications.DELETE("", notifHandler.ClearAll)
		}

		alertRules := protected.Group("/alert-rules")
		{
			alertRules.POST("", notifHandler.CreateRule)
			alertRules.GET("", notifHandler.ListRules)
			alertRules.POST("/:id/toggle", notifHandler.ToggleRule)
			alertRules.DELETE("/:id", notifHandler.DeleteRule)
		}

		imports := protected.Group("/import")
		{
			imports.POST("/preview", importHandler.Preview)
			imports.POST("/csv", importHandler.Import)
			imports.GET("/history", importHandler.ListJobs)
		}

		export := protected.Group("/export")
		{
			export.GET("/transactions", exportHandler.ExportTransactions)
			export.GET("/accounts", exportHandler.ExportAccounts)
			export.GET("/all", exportHandler.ExportAll)
		}

		protected.GET("/exchange-rates", exchangeHandler.GetRates)
		protected.GET("/exchange-rates/convert", exchangeHandler.Convert)

		exchanges := protected.Group("/exchanges")
		{
			exchanges.POST("/connect", exchangeConnHandler.Connect)
			exchanges.GET("", exchangeConnHandler.List)
			exchanges.POST("/:id/sync", exchangeConnHandler.Sync)
			exchanges.POST("/:id/toggle", exchangeConnHandler.Toggle)
			exchanges.DELETE("/:id", exchangeConnHandler.Delete)
		}

		push := protected.Group("/push")
		{
			push.GET("/vapid-key", pushHandler.GetVAPIDKey)
			push.POST("/subscribe", pushHandler.Subscribe)
			push.POST("/unsubscribe", pushHandler.Unsubscribe)
		}

		e2e := protected.Group("/e2e")
		{
			e2e.GET("/export-data", e2eHandler.ExportData)
			e2e.POST("/enable", e2eHandler.MigrateAndEnable)
			e2e.POST("/disable", e2eHandler.MigrateAndDisable)
		}

		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole(models.RoleAdmin))
		{
			admin.GET("/users", adminHandler.ListUsers)
			admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
			admin.GET("/audit-logs", auditHandler.List)
		}
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("MoneyVault API starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Cancel background jobs and wait for them to finish
	cancel()
	wg.Wait()
	log.Println("Background jobs stopped")

	// Gracefully shutdown HTTP server with 10s timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
