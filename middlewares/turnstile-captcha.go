package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type TurnstileResponse struct {
	Success     bool     `json:"success"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
	ChallengeTS string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
}

type TurnstileRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIP string `json:"remoteip,omitempty"`
}

type tokenUsageInfo struct {
	Count     int
	ExpireAt  time.Time
	FirstUsed time.Time
	LastUsed  time.Time
}

type TurnstileMiddleware struct {
	secretKey      string
	enabled        bool
	usedTokens     map[string]*tokenUsageInfo
	mutex          sync.RWMutex
	stopCleanup    chan bool
	httpClient     *http.Client
	verifyEndpoint string
}

const (
	TurnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	TokenTTL           = 5 * time.Minute  // Token 5 dakika geçerli
	CleanupInterval    = 10 * time.Minute // Her 10 dakikada temizlik
	MaxTokenUses       = 5                // Token maksimum 5 kere kullanılabilir
)

func NewTurnstileMiddleware() *TurnstileMiddleware {
	secretKey := os.Getenv("CLOUDFLARE_TURNSTILE_SECRET_KEY")

	middleware := &TurnstileMiddleware{
		secretKey:      secretKey,
		enabled:        secretKey != "",
		usedTokens:     make(map[string]*tokenUsageInfo),
		stopCleanup:    make(chan bool, 1),
		verifyEndpoint: TurnstileVerifyURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				MaxIdleConnsPerHost: 5,
			},
		},
	}

	if !middleware.enabled {
		log.Printf("[TURNSTILE] UYARI: CLOUDFLARE_TURNSTILE_SECRET_KEY boş - middleware devre dışı")
		return middleware
	}

	log.Printf("[TURNSTILE] Middleware etkinleştirildi (Max kullanım: %d, TTL: %v)", MaxTokenUses, TokenTTL)

	go middleware.startCleanupRoutine()

	return middleware
}

func (tm *TurnstileMiddleware) Middleware() gin.HandlerFunc {
	if !tm.enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		token := tm.extractToken(c)
		if token == "" {
			tm.respondWithError(c, http.StatusBadRequest, "turnstile_token_missing",
				"Güvenlik doğrulaması token'ı eksik")
			return
		}

		// Token kullanım durumunu kontrol et
		if tm.isTokenExhausted(token) {
			tm.respondWithError(c, http.StatusBadRequest, "turnstile_token_exhausted",
				fmt.Sprintf("Bu güvenlik token'ı maksimum kullanım sayısına (%d) ulaştı", MaxTokenUses))
			return
		}

		// Token'ı doğrula (sadece ilk kullanımda Cloudflare'e gidiyoruz)
		if !tm.verifyTokenIfNeeded(token, c.ClientIP()) {
			tm.respondWithError(c, http.StatusForbidden, "turnstile_verification_failed",
				"Güvenlik doğrulaması başarısız")
			return
		}

		// Token'ı kullanılmış olarak işaretle
		tm.markTokenAsUsed(token)

		c.Next()
	}
}

func (tm *TurnstileMiddleware) startCleanupRoutine() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			count := tm.cleanupExpiredTokens()
			if count > 0 {
				log.Printf("[TURNSTILE] %d süresi dolmuş token temizlendi", count)
			}
		case <-tm.stopCleanup:
			log.Printf("[TURNSTILE] Token temizleme rutini durduruldu")
			return
		}
	}
}

func (tm *TurnstileMiddleware) extractToken(c *gin.Context) string {
	if token := c.GetHeader("CF-Turnstile-Token"); token != "" {
		return token
	}
	if token := c.GetHeader("X-Turnstile-Token"); token != "" {
		return token
	}
	if token := c.GetHeader("Turnstile-Token"); token != "" {
		return token
	}
	return c.PostForm("cf-turnstile-response")
}

func (tm *TurnstileMiddleware) isTokenExhausted(token string) bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	info, exists := tm.usedTokens[token]
	if !exists {
		return false // Yeni token, henüz kullanılmamış
	}

	now := time.Now()

	// Token süresi dolmuşsa, kullanılmamış sayıyoruz
	if now.After(info.ExpireAt) {
		return false
	}

	// Maksimum kullanım sayısına ulaştı mı?
	return info.Count >= MaxTokenUses
}

func (tm *TurnstileMiddleware) verifyTokenIfNeeded(token, clientIP string) bool {
	tm.mutex.RLock()
	info, exists := tm.usedTokens[token]
	tm.mutex.RUnlock()

	// Eğer token daha önce doğrulanmış ve henüz geçerliyse, tekrar doğrulamaya gerek yok
	if exists && time.Now().Before(info.ExpireAt) {
		log.Printf("[TURNSTILE] Token daha önce doğrulanmış, tekrar doğrulamaya gerek yok (kullanım: %d/%d)", info.Count, MaxTokenUses)
		return true
	}

	// İlk kullanım veya süresi dolmuş, Cloudflare'e doğrulat
	return tm.verifyWithCloudflare(token, clientIP)
}

func (tm *TurnstileMiddleware) verifyWithCloudflare(token, clientIP string) bool {
	requestBody := TurnstileRequest{
		Secret:   tm.secretKey,
		Response: token,
	}

	if clientIP != "" {
		requestBody.RemoteIP = clientIP
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("[TURNSTILE] Request serialize hatası: %v", err)
		return false
	}

	req, err := http.NewRequest("POST", tm.verifyEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[TURNSTILE] Request oluşturma hatası: %v", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "GoTurnstileMiddleware/1.0")

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		log.Printf("[TURNSTILE] API isteği hatası: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[TURNSTILE] API HTTP hatası: %d", resp.StatusCode)
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[TURNSTILE] Response okuma hatası: %v", err)
		return false
	}

	var result TurnstileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[TURNSTILE] JSON parse hatası: %v", err)
		return false
	}

	if !result.Success {
		log.Printf("[TURNSTILE] Doğrulama başarısız - Error codes: %v", result.ErrorCodes)
		return false
	}

	log.Printf("[TURNSTILE] Token başarıyla doğrulandı - Hostname: %s", result.Hostname)
	return true
}

func (tm *TurnstileMiddleware) markTokenAsUsed(token string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	now := time.Now()
	info, exists := tm.usedTokens[token]

	if !exists {
		// İlk kullanım
		tm.usedTokens[token] = &tokenUsageInfo{
			Count:     1,
			ExpireAt:  now.Add(TokenTTL),
			FirstUsed: now,
			LastUsed:  now,
		}
		log.Printf("[TURNSTILE] Token ilk kez kullanıldı (1/%d)", MaxTokenUses)
	} else if now.After(info.ExpireAt) {
		// Süresi dolmuş, yeniden başlat
		tm.usedTokens[token] = &tokenUsageInfo{
			Count:     1,
			ExpireAt:  now.Add(TokenTTL),
			FirstUsed: now,
			LastUsed:  now,
		}
		log.Printf("[TURNSTILE] Süresi dolmuş token yeniden kullanıldı (1/%d)", MaxTokenUses)
	} else {
		// Mevcut kullanımı artır
		info.Count++
		info.LastUsed = now
		log.Printf("[TURNSTILE] Token tekrar kullanıldı (%d/%d)", info.Count, MaxTokenUses)
	}
}

func (tm *TurnstileMiddleware) cleanupExpiredTokens() int {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	now := time.Now()
	cleanedCount := 0

	for token, info := range tm.usedTokens {
		if now.After(info.ExpireAt) {
			delete(tm.usedTokens, token)
			cleanedCount++
		}
	}

	return cleanedCount
}

func (tm *TurnstileMiddleware) respondWithError(c *gin.Context, statusCode int, errorCode, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   errorCode,
		"message": message,
	})
	c.Abort()
}

func (tm *TurnstileMiddleware) Close() {
	if tm.enabled {
		select {
		case tm.stopCleanup <- true:
		default:
		}
	}
}
