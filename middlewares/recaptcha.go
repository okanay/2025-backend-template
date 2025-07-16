package middlewares

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RecaptchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	Score       float64   `json:"score,omitempty"`
	Action      string    `json:"action,omitempty"`
	ErrorCodes  []string  `json:"error-codes,omitempty"`
}

type RecaptchaMiddleware struct {
	secretKey    string
	minimumScore float64
	usedTokens   map[string]bool
	mutex        sync.RWMutex
	stopCleanup  chan bool
}

func NewRecaptchaMiddleware(secretKey string, minimumScore float64) *RecaptchaMiddleware {
	if secretKey == "" {
		log.Println("[RECAPTCHA] UYARI: Secret key boş! Doğrulama devre dışı kalacak.")
	}

	middleware := &RecaptchaMiddleware{
		secretKey:    secretKey,
		minimumScore: minimumScore,
		usedTokens:   make(map[string]bool),
		stopCleanup:  make(chan bool),
	}

	go func() {
		ticker := time.NewTicker(4 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				count := middleware.cleanupUsedTokens()
				log.Printf("[RECAPTCHA] %d kullanılmış token temizlendi", count)
			case <-middleware.stopCleanup:
				log.Println("[RECAPTCHA] Token temizleme rutini durduruldu")
				return
			}
		}
	}()

	return middleware
}

func (rm *RecaptchaMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if rm.secretKey == "" {
			c.Next()
			return
		}

		token := c.GetHeader("X-Recaptcha-Token")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "recaptcha_token_missing",
				"message": "Güvenlik doğrulaması token'ı eksik",
			})
			c.Abort()
			return
		}

		rm.mutex.RLock()
		used := rm.usedTokens[token]
		rm.mutex.RUnlock()

		if used {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "recaptcha_token_used",
				"message": "Bu güvenlik token'ı daha önce kullanılmış (timeout-or-duplicate)",
			})
			c.Abort()
			return
		}

		clientIP := c.ClientIP()

		formData := url.Values{
			"secret":   {rm.secretKey},
			"response": {token},
		}

		if clientIP != "" {
			formData.Add("remoteip", clientIP)
		}

		resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", formData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "recaptcha_verification_failed",
				"message": "Güvenlik doğrulaması yapılamadı: " + err.Error(),
			})
			c.Abort()
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "recaptcha_response_error",
				"message": "Güvenlik doğrulaması yanıtı alınamadı: " + err.Error(),
			})
			c.Abort()
			return
		}

		var result RecaptchaResponse
		if err := json.Unmarshal(body, &result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "recaptcha_parse_error",
				"message": "Güvenlik doğrulaması yanıtı işlenemedi: " + err.Error(),
			})
			c.Abort()
			return
		}

		rm.mutex.Lock()
		rm.usedTokens[token] = true
		rm.mutex.Unlock()

		if !result.Success {
			errorMsg := "Güvenlik doğrulaması başarısız"
			if len(result.ErrorCodes) > 0 {
				switch result.ErrorCodes[0] {
				case "missing-input-secret":
					errorMsg = "Secret parametresi eksik"
				case "invalid-input-secret":
					errorMsg = "Secret parametresi geçersiz veya hatalı biçimlendirilmiş"
				case "missing-input-response":
					errorMsg = "Yanıt parametresi eksik"
				case "invalid-input-response":
					errorMsg = "Yanıt parametresi geçersiz veya hatalı biçimlendirilmiş"
				case "bad-request":
					errorMsg = "İstek geçersiz veya hatalı biçimlendirilmiş"
				case "timeout-or-duplicate":
					errorMsg = "Yanıt artık geçerli değil: ya çok eski ya da daha önce kullanılmış"
				}
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "recaptcha_verification_failed",
				"message": errorMsg,
			})
			c.Abort()
			return
		}

		if result.Score < rm.minimumScore {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "recaptcha_score_low",
				"message": "Güvenlik puanı yetersiz",
				"details": gin.H{
					"score":         result.Score,
					"minimum_score": rm.minimumScore,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rm *RecaptchaMiddleware) cleanupUsedTokens() int {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	count := len(rm.usedTokens)
	rm.usedTokens = make(map[string]bool)
	return count
}

func (rm *RecaptchaMiddleware) Close() {
	close(rm.stopCleanup)
}
