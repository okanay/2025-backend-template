# Goth Service (`services/goth`)

Bu paket, Google, Apple gibi OAuth 2.0 sağlayıcıları üzerinden kullanıcı kimlik doğrulama işlemlerini yöneten bir servis katmanıdır. Popüler `goth` ve `gothic` kütüphanelerini kullanarak, sosyal medya ile giriş süreçlerini basitleştirir ve standartlaştırır.

## Temel Çalışma Prensibi ve Kullandığı Kütüphaneler

Bu servisin temel amacı, farklı sosyal medya sağlayıcılarından gelen çeşitli veri formatlarını, uygulamamızın anlayacağı tek bir standart formata (`types.ProviderUserData`) dönüştürmektir.

1.  **`goth` Kütüphanesi:** Bu kütüphane, `main.go` içinde `SetupGothProviders` fonksiyonu ile yapılandırılır. Hangi sağlayıcıların (Google, Apple vb.) hangi `CLIENT_ID` ve `SECRET_KEY` ile kullanılacağını belirler.

2.  **`gothic` Kütüphanesi:** Bu kütüphane, `gin` ile entegre çalışarak HTTP request'lerini yönetir. Bir kullanıcı `/auth/google` gibi bir adrese gittiğinde, `gothic` devreye girerek kullanıcıyı doğru Google giriş sayfasına yönlendirir. Kullanıcı giriş yapıp geri döndüğünde ise callback'i yakalar ve kullanıcı bilgilerini `goth.User` formatında elde eder.

3.  **`GothService` (Bu Paket):** `gothic`'ten alınan `goth.User` objesini işleyerek, içindeki sağlayıcıya özel bilgileri (kullanıcı ID'si, adı, e-postası vb.) ayıklar ve projemizin genelinde kullanabileceğimiz standart `types.ProviderUserData` yapısına dönüştürür.

## Fonksiyonlar

---

### `HandleProviderCallback`

Tüm sosyal medya sağlayıcılarından gelen callback verilerini işleyen ana fonksiyondur.

-   **Ne Yapar?:** `gothic` tarafından doğrulanmış ve alınmış olan `goth.User` objesini alır ve içindeki bilgileri (Provider, UserID, Email, Name, AvatarURL vb.) standart `*types.ProviderUserData` yapımıza çevirir. Bu sayede `auth` repository'si, hangi sağlayıcıdan gelirse gelsin, her zaman aynı formatta veri ile çalışabilir.
-   **Ne Alır?:** `goth.User` (sağlayıcıdan gelen kullanıcı verisi)
-   **Ne Döndürür?:** `*types.ProviderUserData` (uygulamamızın standart kullanıcı veri formatı).

```go
func (s *Service) HandleProviderCallback(gothUser goth.User) *types.ProviderUserData
```
