package ValidationService

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// APIError, tek bir validasyon hatasını yapısal olarak temsil eder.
// Bu format, frontend'in hataları işlemesini çok kolaylaştırır.
type APIError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// Service, go-playground/validator'ı saran ve özelleştiren ana yapıdır.
type Service struct {
	validate *validator.Validate
}

// NewService, yeni bir validasyon servisi oluşturur ve tüm özel ayarları
// (JSON etiket okuyucu, özel kurallar vb.) tek seferde kaydeder.
func NewService() *Service {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Service{validate: validate}
}

// Validate, gelen isteğin gövdesini (body) bir struct'a atar ve doğrular.
// Hata varsa, HTTP yanıtını otomatik olarak gönderir ve hata listesini döner.
// Hata yoksa, nil döner ve handler akışına devam edebilir.
func (s *Service) Validate(c *gin.Context, req any) []*APIError {
	// 1. Gelen JSON'ı struct'a bind et.
	if err := c.ShouldBindJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_json_format",
			"message": "İsteğin gövdesi geçerli bir JSON formatında değil: " + err.Error(),
		})
		return []*APIError{{Message: "Invalid JSON"}}
	}

	// 2. Struct'ı 'validate' tag'lerine göre doğrula.
	err := s.validate.Struct(req)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "validator_internal_error"})
			return []*APIError{{Message: "Validator internal error"}}
		}

		var apiErrors []*APIError
		for _, fieldErr := range validationErrors {
			apiErrors = append(apiErrors, &APIError{
				Field:   fieldErr.Field(), // RegisterTagNameFunc sayesinde burası artık "firstName" gibi dönecek.
				Tag:     fieldErr.Tag(),
				Message: s.generateErrorMessage(fieldErr),
			})
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"errors":  apiErrors,
		})
		return apiErrors
	}

	return nil
}

func (s *Service) generateErrorMessage(e validator.FieldError) string {
	// e.Field() artık "firstName" gibi `json` etiketindeki adı döndürecektir.
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s alanı zorunludur.", field)
	case "email":
		return fmt.Sprintf("%s alanı geçerli bir e-posta adresi olmalıdır.", field)
	case "min":
		if e.Kind() == reflect.String {
			return fmt.Sprintf("%s alanı en az %s karakter olmalıdır.", field, param)
		}
		return fmt.Sprintf("%s alanı için minimum değer %s'dir.", field, param)
	case "max":
		if e.Kind() == reflect.String {
			return fmt.Sprintf("%s alanı en fazla %s karakter olmalıdır.", field, param)
		}
		return fmt.Sprintf("%s alanı için maksimum değer %s'dir.", field, param)
	case "gte":
		return fmt.Sprintf("%s alanı için minimum değer %s olmalıdır.", field, param)
	case "lte":
		return fmt.Sprintf("%s alanı için maksimum değer %s olmalıdır.", field, param)
	case "oneof":
		return fmt.Sprintf("%s alanı sadece şu değerlerden biri olabilir: [%s].", field, param)

	default:
		return fmt.Sprintf("%s alanı için '%s' kuralı sağlanamadı.", field, tag)
	}
}
