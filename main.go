package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	c "github.com/okanay/backend-template/configs"
	db "github.com/okanay/backend-template/database"
	fh "github.com/okanay/backend-template/handlers/file"
	gfm "github.com/okanay/backend-template/handlers/github-file-manager"
	mh "github.com/okanay/backend-template/handlers/globals"
	uh "github.com/okanay/backend-template/handlers/user"

	"github.com/okanay/backend-template/middlewares"
	mw "github.com/okanay/backend-template/middlewares"
	fr "github.com/okanay/backend-template/repositories/file"
	tr "github.com/okanay/backend-template/repositories/token"
	ur "github.com/okanay/backend-template/repositories/user"
	gr "github.com/okanay/backend-template/services/github"
	GothService "github.com/okanay/backend-template/services/goth"
	r2r "github.com/okanay/backend-template/services/r2"

	"github.com/okanay/backend-template/services/cache"
)

func main() {
	// 1. Çevresel Değişkenleri Yükle
	if err := godotenv.Load(".env"); err != nil {
		log.Println("[ENV]: .env file could not be loaded, environment variables will be used")
	} else {
		log.Println("[ENV]: .env file loaded successfully.")
	}

	// 2. Veritabanı Bağlantısı Kur
	sqlDB, err := db.Init(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("[DATABASE]: Error connecting to the database: %v", err)
	} else {
		log.Println("[DATABASE]: Database connection established successfully")
		defer sqlDB.Close()
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// 3. Servisleri ve Handler'ları Başlat
	repos := initRepositories(sqlDB)
	services := initServices()
	handlers := initHandlers(repos, services)

	// 4. Router ve Middleware Yapılandırması
	router := gin.Default()
	// configs.SetupTrustedProxies(router)
	router.Use(c.CorsConfig())
	router.Use(c.SecureConfig)
	router.MaxMultipartMemory = 10 << 20

	// 4.0 Global Middlewares
	router.Use(middlewares.TimeoutMiddleware())

	// 4.1 Turnstile Middleware
	turnstileMiddleware := middlewares.NewTurnstileMiddleware()
	defer turnstileMiddleware.Close()

	// 4.2 Route Grupları
	publicAPI := router.Group("/public")
	publicFileAPI := router.Group("/public/files")
	internalAPI := router.Group("/internal")
	authAPI := router.Group("/auth")

	// 4.3 Grup Bazlı Middleware'ler
	publicAPI.Use(mw.RateLimiterMiddleware(400, time.Minute))
	internalAPI.Use(mw.RateLimiterMiddleware(1000, time.Minute))

	authAPI.Use(mw.RateLimiterMiddleware(1000, time.Minute))
	authAPI.Use(mw.AuthMiddleware(repos.User, repos.Token))

	publicFileAPI.Use(mw.RateLimiterMiddleware(10, 120*time.Minute))

	// 4.4 Ana ve NotFound Route
	router.GET("/", handlers.Main.Index)
	router.NoRoute(handlers.Main.NotFound)

	// 4.5 Route Organizasyonu

	// --- PUBLIC AUTH ---
	publicAPI.POST("/login", handlers.User.Login)
	publicAPI.POST("/register", handlers.User.Register)
	publicAPI.GET("/test-ip-address", handlers.Main.IPTestEndpoint)

	// --- OAUTH PROVIDER AUTH ---
	// Gothic middleware'i ekleyebilirsiniz (isteğe bağlı)
	oauthAPI := publicAPI.Group("/auth")
	// oauthAPI.Use(middlewares.GothicMiddleware()) // İsteğe bağlı
	oauthAPI.GET("/:provider", handlers.User.ProviderHandler)          // Provider'a yönlendirme
	oauthAPI.GET("/:provider/callback", handlers.User.CallbackHandler) // Provider callback

	// --- PUBLIC FILES ---
	publicFileAPI.Use(turnstileMiddleware.Middleware())
	publicFileAPI.POST("/presigned-url", handlers.File.CreatePresignedURL)
	publicFileAPI.POST("/confirm-upload", handlers.File.ConfirmUpload)

	// --- AUTH USER ---
	authAPI.GET("/logout", handlers.User.Logout)
	authAPI.GET("/get-me", handlers.User.GetMe)

	// --- AUTH FILES ---
	authAPI.GET("/files/category", handlers.File.GetFilesByCategory)
	authAPI.POST("/files/presigned-url", handlers.File.CreatePresignedURL)
	authAPI.POST("/files/confirm-upload", handlers.File.ConfirmUpload)
	authAPI.DELETE("/files/:id", handlers.File.DeleteFile)

	// --- AUTH DOCUMENTS ---
	authAPI.GET("/content/categories", handlers.GithubFileManager.GetCategories)
	authAPI.GET("/content/:category", handlers.GithubFileManager.GetContent)
	authAPI.POST("/content/:category/save", handlers.GithubFileManager.SaveContent)
	authAPI.GET("/content/:category/draft-status", handlers.GithubFileManager.GetDraftStatus)
	authAPI.POST("/content/:category/publish", handlers.GithubFileManager.PublishCategory)
	authAPI.DELETE("/content/:category/restart", handlers.GithubFileManager.RestartCategory)

	// 5. Sunucuyu Başlat
	startServer(router)
}

type Repositories struct {
	User  *ur.Repository
	Token *tr.Repository
	File  *fr.Repository
}

type Services struct {
	Goth   *GothService.Service
	Cache  cache.CacheService
	R2     *r2r.Service
	Github *gr.Service
}

type Handlers struct {
	Main              *mh.Handler
	User              *uh.Handler
	File              *fh.Handler
	GithubFileManager *gfm.Handler
}

// Repository'lerin başlatılması
func initRepositories(sqlDB *sql.DB) Repositories {
	return Repositories{
		User:  ur.NewRepository(sqlDB),
		Token: tr.NewRepository(sqlDB),
		File:  fr.NewRepository(sqlDB),
	}
}

// initServices fonksiyonunu da güncelle
func initServices() Services {
	GothService.SetupGothProviders()

	return Services{
		Goth:   GothService.NewService(),
		Cache:  cache.NewCacheService(1 * time.Hour),
		Github: gr.NewService(os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPOSITORY_NAME"), os.Getenv("GITHUB_TOKEN")),
		R2: r2r.NewService(
			os.Getenv("R2_ACCOUNT_ID"),
			os.Getenv("R2_ACCESS_KEY_ID"),
			os.Getenv("R2_ACCESS_KEY_SECRET"),
			os.Getenv("R2_BUCKET_NAME"),
			os.Getenv("R2_FOLDER_NAME"),
			os.Getenv("R2_PUBLIC_URL_BASE"),
			os.Getenv("R2_ENDPOINT"),
		),
	}
}

// Handler'ların başlatılması
func initHandlers(repos Repositories, services Services) Handlers {
	return Handlers{
		Main:              mh.NewHandler(),
		User:              uh.NewHandler(services.Goth, repos.User, repos.Token),
		File:              fh.NewHandler(repos.File, services.R2),
		GithubFileManager: gfm.NewHandler(services.Github),
	}
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
