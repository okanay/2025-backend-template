package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/okanay/backend-template/configs"
	db "github.com/okanay/backend-template/database"
	AuthHandler "github.com/okanay/backend-template/handlers/auth"
	FileHandler "github.com/okanay/backend-template/handlers/file"
	GithubHandler "github.com/okanay/backend-template/handlers/github"
	StaticRoutesHandler "github.com/okanay/backend-template/handlers/static-route-handlers"
	AuthRepository "github.com/okanay/backend-template/repositories/auth"
	FileRepository "github.com/okanay/backend-template/repositories/file"
	TokenRepository "github.com/okanay/backend-template/repositories/token"

	"github.com/okanay/backend-template/middlewares"

	AutomationService "github.com/okanay/backend-template/services/automation"
	cache "github.com/okanay/backend-template/services/cache"
	GithubService "github.com/okanay/backend-template/services/github"
	GothService "github.com/okanay/backend-template/services/goth"
	R2Service "github.com/okanay/backend-template/services/r2"
	ValidationService "github.com/okanay/backend-template/services/validation"
)

func main() {
	// 1. Uygulama Yapılandırmasını Yükle
	loadConfig()

	// 2. Veritabanı Bağlantısını Kur
	db := setupDatabase()
	defer db.Close()

	// 3. Gin Router'ını Başlat
	router := setupRouter()

	// Repositories (Veritabanı Erişim Katmanı)
	userRepo := AuthRepository.NewRepository(db)
	tokenRepo := TokenRepository.NewRepository(db)
	fileRepo := FileRepository.NewRepository(db)

	// Services Initialization
	AutomationService := AutomationService.NewService()
	ValidationService := ValidationService.NewService()

	GothService.SetupGothProviders()
	CacheService := cache.NewCacheService(1 * time.Hour)

	// Services (İş Mantığı Katmanı)
	gothService := GothService.NewService()
	githubService := GithubService.NewService(
		os.Getenv("GITHUB_OWNER"),
		os.Getenv("GITHUB_REPOSITORY_NAME"),
		os.Getenv("GITHUB_TOKEN"),
	)
	r2Service := R2Service.NewService(
		os.Getenv("R2_ACCOUNT_ID"),
		os.Getenv("R2_ACCESS_KEY_ID"),
		os.Getenv("R2_ACCESS_KEY_SECRET"),
		os.Getenv("R2_BUCKET_NAME"),
		os.Getenv("R2_FOLDER_NAME"),
		os.Getenv("R2_PUBLIC_URL_BASE"),
		os.Getenv("R2_ENDPOINT"),
	)

	// Handlers (İstekleri İşleyen Katman)
	staticHandler := StaticRoutesHandler.NewHandler(ValidationService)
	authHandler := AuthHandler.NewHandler(gothService, userRepo, tokenRepo, ValidationService)
	fileHandler := FileHandler.NewHandler(fileRepo, r2Service, ValidationService)
	githubHandler := GithubHandler.NewHandler(githubService, ValidationService)

	// --- AUTOMATION ---

	AutomationService.Add("cache:cleanup-expired", "@every 1h", func() {
		if inMemoryCache, ok := CacheService.(*cache.InMemoryCache); ok {
			inMemoryCache.Cleanup()
		}
	})

	// --- ROTALAR ---

	// Global ve 404 Rotaları
	router.GET("/", staticHandler.Index)
	router.NoRoute(staticHandler.NotFound)
	router.POST("/test-validate", staticHandler.TestValidate)

	// API Versiyonlama Grubu (v1)
	v1 := router.Group("/v1")
	{
		// --- Public Rotalar (Kimlik Doğrulama GEREKTİRMEYEN) ---
		public := v1.Group("/")
		public.Use(middlewares.RateLimiterMiddleware(400, time.Minute))
		{
			// Kullanıcı Kimlik Doğrulama
			public.POST("/auth/register", authHandler.Register)
			public.POST("/auth/login", authHandler.Login)

			// Sosyal Medya (OAuth)
			public.GET("/auth/provider/:provider", authHandler.ProviderHandler)
			public.GET("/auth/provider/:provider/callback", authHandler.CallbackHandler)
		}

		// --- Protected Rotalar (Kimlik Doğrulama GEREKTİREN) ---
		protected := v1.Group("/")
		protected.Use(middlewares.RateLimiterMiddleware(1000, time.Minute))
		protected.Use(middlewares.AuthMiddleware(userRepo, tokenRepo))
		protected.Use(middlewares.PermissionMiddleware(CacheService, userRepo))
		{
			test := protected.Group("/test")
			{
				test.GET("/ip-address", staticHandler.IPTestEndpoint)
			}

			// Oturum ve Kullanıcı
			protected.GET("/auth/me", authHandler.GetMe)
			protected.POST("/auth/logout", authHandler.Logout)

			// Dosya Yönetimi
			protected.GET("/files", fileHandler.GetFilesByCategory)
			protected.DELETE("/files/:id", fileHandler.DeleteFile)
			protected.POST("/files/presigned-url", fileHandler.CreatePresignedURL)
			protected.POST("/files/confirm-upload", fileHandler.ConfirmUpload)

			// İçerik Yönetimi (GitHub)
			content := protected.Group("/github")
			{
				content.GET("/categories", githubHandler.GetCategories)
				content.GET("/:category", githubHandler.GetContent)
				content.POST("/:category/save", githubHandler.SaveContent)
				content.GET("/:category/draft-status", githubHandler.GetDraftStatus)
				content.POST("/:category/publish", githubHandler.PublishCategory)
				content.DELETE("/:category/restart", githubHandler.RestartCategory)
			}
		}
	}

	startServer(router)
}

// loadConfig, .env dosyasını yükler.
func loadConfig() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("[ENV]: .env file could not be loaded, environment variables will be used")
	} else {
		log.Println("[ENV]: .env file loaded successfully.")
	}
}

// setupDatabase, veritabanı bağlantısını kurar ve ayarlar.
func setupDatabase() *sql.DB {
	db, err := db.Init(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("[DATABASE]: Error connecting to the database: %v", err)
	}
	log.Println("[DATABASE]: Database connection established successfully")

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db
}

// setupRouter, Gin router'ını oluşturur ve global middleware'leri ekler.
func setupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(configs.CorsConfig())
	router.Use(configs.SecureConfig)
	router.Use(middlewares.TimeoutMiddleware())
	router.MaxMultipartMemory = 10 << 20
	return router
}

// Sunucuyu başlatır
func startServer(router *gin.Engine) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[SERVER]: %s portu üzerinde dinleniyor...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[SERVER]: Sunucu başlatılırken hata: %v", err)
	}
}
