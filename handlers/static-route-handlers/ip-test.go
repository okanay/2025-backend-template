package StaticRoutesHandler

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) IPTestEndpoint(c *gin.Context) {
	// Mevcut durumda ne geldiğini göster
	clientIP := c.ClientIP()

	// Tüm Cloudflare header bilgilerini topla
	headers := map[string]string{
		// IP Address Headers
		"CF-Connecting-IP":   c.Request.Header.Get("CF-Connecting-IP"),
		"CF-Connecting-IPv6": c.Request.Header.Get("CF-Connecting-IPv6"),
		"True-Client-IP":     c.Request.Header.Get("True-Client-IP"),
		"X-Forwarded-For":    c.Request.Header.Get("X-Forwarded-For"),
		"X-Real-IP":          c.Request.Header.Get("X-Real-IP"),

		// Protocol and Connection Headers
		"X-Forwarded-Proto": c.Request.Header.Get("X-Forwarded-Proto"),
		"CF-Visitor":        c.Request.Header.Get("CF-Visitor"),
		"Connection":        c.Request.Header.Get("Connection"),

		// Cloudflare Specific Headers
		"Cf-Ray":         c.Request.Header.Get("Cf-Ray"),
		"CF-IPCountry":   c.Request.Header.Get("CF-IPCountry"),
		"CF-Worker":      c.Request.Header.Get("CF-Worker"),
		"CF-EW-Via":      c.Request.Header.Get("CF-EW-Via"),
		"CF-Pseudo-IPv4": c.Request.Header.Get("CF-Pseudo-IPv4"),
		"CDN-Loop":       c.Request.Header.Get("CDN-Loop"),

		// Encoding Headers
		"Accept-Encoding": c.Request.Header.Get("Accept-Encoding"),

		// Cache Headers
		"Cf-Cache-Status": c.Request.Header.Get("Cf-Cache-Status"),

		// Authentication & Session
		"Cookie": c.Request.Header.Get("Cookie"),

		// Standard Headers
		"User-Agent":      c.Request.Header.Get("User-Agent"),
		"Accept":          c.Request.Header.Get("Accept"),
		"Accept-Language": c.Request.Header.Get("Accept-Language"),
		"Referer":         c.Request.Header.Get("Referer"),
		"Origin":          c.Request.Header.Get("Origin"),

		// Security Headers
		"X-Requested-With": c.Request.Header.Get("X-Requested-With"),
		"Sec-Fetch-Site":   c.Request.Header.Get("Sec-Fetch-Site"),
		"Sec-Fetch-Mode":   c.Request.Header.Get("Sec-Fetch-Mode"),
		"Sec-Fetch-Dest":   c.Request.Header.Get("Sec-Fetch-Dest"),
	}

	// Boş header'ları temizle
	cleanedHeaders := make(map[string]string)
	for key, value := range headers {
		if strings.TrimSpace(value) != "" {
			cleanedHeaders[key] = value
		}
	}

	// CF-Visitor header'ını parse et (JSON formatında gelir)
	var cfVisitorData map[string]interface{}
	cfVisitorRaw := c.Request.Header.Get("CF-Visitor")
	if cfVisitorRaw != "" {
		json.Unmarshal([]byte(cfVisitorRaw), &cfVisitorData)
	}

	// IP öncelik sırası belirleme
	ipPriority := []string{
		c.Request.Header.Get("CF-Connecting-IP"),
		c.Request.Header.Get("True-Client-IP"),
		c.Request.Header.Get("X-Real-IP"),
		c.Request.Header.Get("X-Forwarded-For"),
	}

	var realClientIP string
	for _, ip := range ipPriority {
		if strings.TrimSpace(ip) != "" {
			// X-Forwarded-For birden fazla IP içerebilir, ilkini al
			if strings.Contains(ip, ",") {
				realClientIP = strings.TrimSpace(strings.Split(ip, ",")[0])
			} else {
				realClientIP = strings.TrimSpace(ip)
			}
			break
		}
	}

	// Rate limiter'da kullanılan IP'yi de göster
	rateLimitIP := c.ClientIP()

	// Ülke bilgisi analizi
	countryCode := c.Request.Header.Get("CF-IPCountry")
	var countryInfo map[string]string
	switch countryCode {
	case "TR":
		countryInfo = map[string]string{
			"code": "TR",
			"name": "Turkey",
			"note": "Türkiye'den gelen istek",
		}
	case "XX":
		countryInfo = map[string]string{
			"code": "XX",
			"name": "Unknown",
			"note": "Ülke bilgisi bulunamadı",
		}
	case "T1":
		countryInfo = map[string]string{
			"code": "T1",
			"name": "Tor Network",
			"note": "Tor ağından gelen istek - güvenlik riski!",
		}
	default:
		if countryCode != "" {
			countryInfo = map[string]string{
				"code": countryCode,
				"name": "Detected Country",
				"note": "Cloudflare tarafından tespit edilen ülke",
			}
		}
	}

	// Protocol analizi
	var protocolInfo map[string]interface{}
	if cfVisitorData != nil {
		protocolInfo = map[string]interface{}{
			"scheme":     cfVisitorData["scheme"],
			"is_https":   cfVisitorData["scheme"] == "https",
			"raw_data":   cfVisitorData,
			"header_raw": cfVisitorRaw,
		}
	}

	response := gin.H{
		"success": true,
		"data": gin.H{
			"message":        "IP ve Cloudflare bilgileri başarıyla alındı",
			"client_ip":      clientIP,
			"real_client_ip": realClientIP,
			"rate_limit_ip":  rateLimitIP,
			"is_same":        clientIP == rateLimitIP,
			"ip_analysis": gin.H{
				"gin_detected":    clientIP,
				"cloudflare_real": realClientIP,
				"priority_source": func() string {
					if c.Request.Header.Get("CF-Connecting-IP") != "" {
						return "CF-Connecting-IP"
					}
					if c.Request.Header.Get("True-Client-IP") != "" {
						return "True-Client-IP"
					}
					if c.Request.Header.Get("X-Real-IP") != "" {
						return "X-Real-IP"
					}
					if c.Request.Header.Get("X-Forwarded-For") != "" {
						return "X-Forwarded-For"
					}
					return "Gin Default"
				}(),
				"matches": clientIP == realClientIP,
			},
			"cloudflare_info": gin.H{
				"ray_id":      c.Request.Header.Get("Cf-Ray"),
				"country":     countryInfo,
				"protocol":    protocolInfo,
				"worker_zone": c.Request.Header.Get("CF-Worker"),
				"is_ipv6":     c.Request.Header.Get("CF-Connecting-IPv6") != "",
				"pseudo_ipv4": c.Request.Header.Get("CF-Pseudo-IPv4"),
				"cdn_loop":    c.Request.Header.Get("CDN-Loop"),
			},
			"security_info": gin.H{
				"is_tor":         countryCode == "T1",
				"fetch_site":     c.Request.Header.Get("Sec-Fetch-Site"),
				"fetch_mode":     c.Request.Header.Get("Sec-Fetch-Mode"),
				"fetch_dest":     c.Request.Header.Get("Sec-Fetch-Dest"),
				"requested_with": c.Request.Header.Get("X-Requested-With"),
			},
			"headers":   cleanedHeaders,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	c.JSON(200, response)
}
