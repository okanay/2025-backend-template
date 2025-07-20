# Token Repository (`repositories/token`)

Bu paket, kullanıcı oturumlarının güvenliğini ve sürekliliğini sağlayan "refresh token" mekanizmasının veritabanı operasyonlarından sorumludur.

## Temel Çalışma Prensibi

Kullanıcılar giriş yaptığında, onlara iki tür token verilir:

1.  **Access Token (JWT):** Kısa ömürlü (örn: 5-15 dakika), kullanıcının kimliğini ve yetkilerini içeren bir token. Her API isteğinde bu token kullanılır.
2.  **Refresh Token:** Uzun ömürlü (örn: 30 gün), sadece yeni bir `Access Token` almak için kullanılan özel bir token. Bu token, veritabanının `refresh_tokens` tablosunda saklanır.

`Access Token`'ın süresi dolduğunda, backend `Refresh Token`'ı kullanarak yeni bir `Access Token` üretir ve kullanıcı oturumunu kesintisiz devam ettirir. Bu repository, `refresh_tokens` tablosundaki bu kayıtların yönetimini yapar.

## Fonksiyonlar

Bu repository'deki ana fonksiyonlar ve görevleri aşağıda açıklanmıştır.

---

### `CreateRefreshToken`

Yeni bir refresh token kaydını veritabanına ekler.

-   **Ne Yapar?:** Kullanıcı giriş yaptığında veya sosyal medya ile bağlandığında, o oturuma özel yeni bir refresh token kaydını `refresh_tokens` tablosuna oluşturur.
-   **Ne Alır?:** `context`, `types.TokenCreateRequest` (kullanıcı ID'si, token string'i, IP adresi, son kullanma tarihi vb.)
-   **Ne Döndürür?:** Oluşturulan `*types.RefreshToken` kaydının tam hali ve `error`.

```go
func (r *Repository) CreateRefreshToken(ctx context.Context, request types.TokenCreateRequest) (*types.RefreshToken, error)
```

---

### `SelectRefreshTokenByToken`

Verilen token string'ine göre geçerli bir refresh token kaydını veritabanından bulur.

-   **Ne Yapar?:** `refresh_tokens` tablosunda, süresi dolmamış (`expires_at > NOW()`) ve iptal edilmemiş (`is_revoked = FALSE`) bir token'ı arar. Yeni bir `Access Token` üretmeden önce bu kontrol yapılır.
-   **Ne Alır?:** `context`, `string` (refresh token)
-   **Ne Döndürür?:** `*types.RefreshToken` ve `error`.

```go
func (r *Repository) SelectRefreshTokenByToken(ctx context.Context, tokenStr string) (*types.RefreshToken, error)
```

---

### `RevokeRefreshToken`

Belirli bir refresh token'ı iptal eder (geçersiz kılar).

-   **Ne Yapar?:** Kullanıcı çıkış yaptığında (`logout`) veya şüpheli bir durumda, ilgili token kaydının `is_revoked` alanını `TRUE` olarak günceller. Bu token artık yeni bir `Access Token` almak için kullanılamaz.
-   **Ne Alır?:** `context`, `string` (iptal edilecek token), `string` (iptal sebebi)
-   **Ne Döndürür?:** `error`.

```go
func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenStr string, reason string) error
```

---

### `RevokeAllUserTokens`

Bir kullanıcıya ait tüm aktif refresh token'ları iptal eder.

-   **Ne Yapar?:** "Tüm cihazlardan çıkış yap" gibi senaryolarda kullanılır. Belirtilen `userID`'ye ait tüm geçerli refresh token'ları `is_revoked = TRUE` olarak günceller.
-   **Ne Alır?:** `context`, `uuid.UUID` (kullanıcı ID'si), `string` (iptal sebebi)
-   **Ne Döndürür?:** İptal edilen token sayısı (`int64`) ve `error`.

```go
func (r *Repository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, reason string) (int64, error)
```

---

### `UpdateRefreshTokenLastUsed`

Bir refresh token'ın son kullanım zamanını günceller.

-   **Ne Yapar?:** Bir refresh token, yeni bir access token almak için kullanıldığında, `last_used_at` alanını güncelleyerek hangi oturumun en son ne zaman aktif olduğunu takip eder.
-   **Ne Alır?:** `context`, `string` (kullanılan token)
-   **Ne Döndürür?:** `error`.

```go
func (r *Repository) UpdateRefreshTokenLastUsed(ctx context.Context, tokenStr string) error
```

## Önemli Notlar

-   **UUID Üretimi:** Bu repository'de oluşturulan tüm `refresh_tokens` kayıtlarının `ID`'leri, veritabanı yerine Go backend'inde `uuid.NewV7()` ile üretilir.
-   **Güvenlik:** Refresh token'lar hassas verilerdir. Asla client-side script'lerin erişebileceği yerlerde (örneğin `localStorage`) saklanmamalıdır. Genellikle `HttpOnly` ve `Secure` cookie'ler içinde saklanırlar.
