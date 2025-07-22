package StaticRoutesHandler

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// --- Veri Modelleri (Structs) ---

// ResponseData, API'nin döndüreceği ana veri yapısıdır.
type ResponseData struct {
	Message      string            `json:"message"`
	Timestamp    string            `json:"timestamp"`
	IPInfo       IPInfo            `json:"ip_info"`
	Cloudflare   CloudflareInfo    `json:"cloudflare_info,omitempty"`
	RelevantHdrs map[string]string `json:"relevant_headers,omitempty"`
}

// IPInfo, IP adresi ile ilgili analiz sonuçlarını içerir.
type IPInfo struct {
	RealClientIP string `json:"real_client_ip"`
	ConfigSource string `json:"config_source"`
	GINClientIP  string `json:"gin_client_ip"`
}

// CloudflareInfo, Cloudflare'e özgü önemli bilgileri içerir.
type CloudflareInfo struct {
	RayID       string `json:"ray_id,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	IsHTTPS     bool   `json:"is_https"`
}

// --- Ana Handler Fonksiyonu ---

func (h *Handler) IPTestEndpoint(c *gin.Context) {
	realIP, source := getRealClientIP(c)

	response := ResponseData{
		Message:   "IP ve Cloudflare bilgileri başarıyla alındı",
		Timestamp: time.Now().Format(time.RFC3339),
		IPInfo: IPInfo{
			RealClientIP: realIP,
			ConfigSource: source,
			GINClientIP:  c.ClientIP(),
		},
		Cloudflare:   getCloudflareInfo(c),
		RelevantHdrs: getRelevantHeaders(c),
	}

	c.JSON(200, gin.H{"success": true, "data": response})
}

// --- Yardımcı Fonksiyonlar (Helpers) ---

// getRealClientIP, çeşitli başlıkları kontrol ederek en güvenilir istemci IP'sini bulur.
func getRealClientIP(c *gin.Context) (ip, source string) {
	// Özel başlığımız her zaman en yüksek önceliğe sahiptir.
	if val := c.Request.Header.Get("X-True-Client-IP"); val != "" {
		return val, "X-True-Client-IP"
	}
	// Standart Cloudflare başlığı ikinci önceliktir.
	if val := c.Request.Header.Get("CF-Connecting-IP"); val != "" {
		return val, "CF-Connecting-IP"
	}
	// Yedek olarak X-Forwarded-For (ilk IP'yi alır).
	if val := c.Request.Header.Get("X-Forwarded-For"); val != "" {
		return strings.TrimSpace(strings.Split(val, ",")[0]), "X-Forwarded-For"
	}
	// Hiçbiri yoksa, Gin'in varsayılanını kullanır.
	return c.ClientIP(), "Gin Default"
}

// getCloudflareInfo, Cloudflare başlıklarından temel bilgileri ayıklar.
func getCloudflareInfo(c *gin.Context) CloudflareInfo {
	var isHTTPS bool
	if visitor := c.Request.Header.Get("CF-Visitor"); visitor != "" {
		var data map[string]string
		if json.Unmarshal([]byte(visitor), &data) == nil {
			isHTTPS = data["scheme"] == "https"
		}
	}

	return CloudflareInfo{
		RayID:       c.Request.Header.Get("Cf-Ray"),
		CountryCode: c.Request.Header.Get("CF-IPCountry"),
		IsHTTPS:     isHTTPS,
	}
}

// getRelevantHeaders, sadece hata ayıklama için en önemli olan başlıkları toplar.
func getRelevantHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	relevant := []string{
		"CF-Connecting-IP",
		"X-True-Client-IP",
		"X-Forwarded-For",
		"CF-IPCountry",
		"CF-Visitor",
		"User-Agent",
		"Cf-Ray",
	}

	for _, key := range relevant {
		if val := c.Request.Header.Get(key); val != "" {
			headers[key] = val
		}
	}
	return headers
}
