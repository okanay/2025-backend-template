# R2 Service (`services/r2`)

Bu paket, Cloudflare R2 gibi S3-uyumlu object storage servisleri ile etkileşim kurmak için bir soyutlama katmanı (abstraction layer) görevi görür. Dosya yüklemek için güvenli URL'ler oluşturma ve mevcut dosyaları silme gibi operasyonları yönetir. AWS SDK for Go V2'yi kullanarak R2 ile iletişim kurar.

## Temel Çalışma Prensibi

Bu servis, doğrudan dosya yükleme veya silme işlemi yapmaz. Bunun yerine, bu işlemleri güvenli bir şekilde gerçekleştirmek için gerekli olan komutları ve URL'leri hazırlar.

-   **Dosya Yükleme:** Frontend'in bir dosyayı doğrudan ve güvenli bir şekilde R2'ye yükleyebilmesi için, süresi kısıtlı ve tek kullanımlık bir "ön-imzalı URL" (presigned URL) oluşturur. Bu, büyük dosyaların backend sunucusu üzerinden geçmesini engelleyerek performansı artırır ve sunucu yükünü azaltır.
-   **Dosya İsimlendirme:** Güvenlik ve çakışmaları önlemek amacıyla, yüklenen dosyaların adları yeniden oluşturulur. Orijinal dosya adı sanitize edilir (güvenli olmayan karakterler kaldırılır) ve sonuna rastgele bir hash eklenir. Örnek: `my-document.pdf` -> `my-document-a1b2c3d4.pdf`.

## Fonksiyonlar

---

### `GeneratePresignedURL`

Frontend'in bir dosyayı R2'ye yükleyebilmesi için güvenli ve süresi kısıtlı bir URL oluşturur.

-   **Ne Yapar?:** AWS SDK'sını kullanarak, belirtilen dosya adı, tipi ve boyutu için bir `PUT` isteği yapmaya izin veren bir URL üretir. Bu URL yaklaşık 5 dakika geçerlidir. Ayrıca, dosyanın R2'deki nihai public URL'ini de oluşturur.
-   **Ne Alır?:** `context`, `types.PresignURLInput` (yüklenecek dosyanın adı, tipi, kategorisi ve boyutu)
-   **Ne Döndürür?:** `*types.PresignedURLOutput` (içinde `PresignedURL` ve nihai `UploadURL` bulunan bir struct) ve `error`.

```go
func (r *Service) GeneratePresignedURL(ctx context.Context, input types.PresignURLInput) (*types.PresignedURLOutput, error)
```

---

### `DeleteObject`

R2 bucket'ından bir nesneyi (dosyayı) siler.

-   **Ne Yapar?:** Verilen `objectKey` (dosyanın R2'deki tam yolu, örn: `uploads/general/my-document-a1b2c3d4.pdf`) için R2'ye bir silme komutu gönderir. Bu işlem, genellikle `file` repository'sindeki `DeleteFileByID` ile birlikte çağrılır.
-   **Ne Alır?:** `context`, `string` (silinecek nesnenin anahtarı)
-   **Ne Döndürür?:** `error`.

```go
func (r *Service) DeleteObject(ctx context.Context, objectKey string) error
```
