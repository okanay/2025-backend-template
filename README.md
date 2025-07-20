# 🚀 Go Backend Şablonu

Modern, ölçeklenebilir ve güvenli web uygulamaları ile API'ler geliştirmek için tasarlanmış **üretime hazır** Go backend şablonu. Kanıtlanmış mimari desenleri ve en iyi pratikleri kullanarak yeni projelere hızlı başlangıç sağlar.

## 🏗️ Mimari Felsefesi

Bu proje, **Temiz Mimari** (Clean Architecture) prensiplerinden ilham alan **Katmanlı Mimari** (Layered Architecture) üzerine kurulmuştur.

### 🎯 Temel Hedefler

- **🔄 Sorumlulukların Ayrılığı:** Her katmanın (sunum, iş mantığı, veri erişim) net ve tek sorumluluğu
- **⬆️ Bağımlılık Kuralı:** Bağımlılıklar daima dış → iç katmanlara doğrudur
- **♻️ Yeniden Kullanılabilirlik:** Merkezi iş mantığı farklı sunum katmanları tarafından kullanılabilir

## ✨ Öne Çıkan Özellikler

### 🔐 Gelişmiş Kimlik Doğrulama
- JWT tabanlı Access/Refresh Token mekanizması
- HttpOnly cookie'ler ile güvenli oturum yönetimi

### 🌐 Sosyal Giriş (OAuth)
- Goth kütüphanesi ile Google, GitHub entegrasyonu
- Kolayca genişletilebilir sosyal giriş altyapısı

### 🛡️ Granüler Yetkilendirme
- **Rol tabanlı:** Admin, User gibi basit roller
- **Aksiyon bazlı:** `posts:create`, `files:upload` gibi detaylı izinler

### 📁 Güvenli Dosya Yükleme
- Cloudflare R2 için ön-imzalı URL (presigned URL)
- Dosyalar doğrudan istemciden R2'ye yüklenir
- Sunucu darboğaz olmaktan çıkar

### ⏰ Asenkron & Zamanlanmış Görevler
- Periyodik işlemler (günlük raporlama, veri temizleme)
- Kilit mekanizmalı otomasyon servisi

### 🗄️ Veritabanı Migrasyonları
- `golang-migrate` ile SQL tabanlı migrasyon
- Versiyon kontrollü şema yönetimi

### ⚡ Performans Odaklı Caching
- Redis entegrasyonu
- Sık erişilen verilerin önbelleğe alınması

### 🔄 CI/CD Entegrasyonu
- GitHub Actions ile otomatik test, build ve deployment
- Hazır iş akışları

### 🌍 Edge Logic (Cloudflare Workers)
- R2 varlıklarının sunulması
- Bot koruması ve A/B testi
- Edge'de çalışan TypeScript mantığı

### 🔒 Güvenlik Odaklı Middlewares
- Rate Limiting (hız sınırlama)
- Timeout koruması
- CORS yapılandırması
- Cloudflare Turnstile bot koruması

## 🛠️ Teknoloji Mimarisi

| Kategori | Teknoloji |
|----------|-----------|
| **Dil & Framework** | Go, Gin Web Framework |
| **Veritabanı** | PostgreSQL |
| **Cache** | Redis |
| **Dosya Depolama** | Cloudflare R2 |
| **Deployment** | Hetzner Ubuntu VPS, Nginx |
| **CI/CD** | GitHub Actions |
| **Edge & DNS** | Cloudflare (Workers, R2) |
| **Auth** | JWT, Goth (OAuth 2.0) |

## 📂 Proje Yapısı

```
├── cmd/                    # 🎯 Derlenebilir ana giriş noktaları
│   └── migrate/           # Database migration CLI aracı
├── configs/               # ⚙️ Sabitler ve statik konfigürasyonlar
├── database/              # 🗄️ Veritabanı bağlantısı ve migrasyonlar
│   └── migrations/        # SQL migrasyon dosyaları
├── handlers/              # 🌐 Sunum Katmanı (HTTP handlers)
├── middlewares/           # 🔗 Gin ara katmanları
├── repositories/          # 💾 Veri Erişim Katmanı
├── services/              # 🧠 İş Mantığı Katmanı
├── types/                 # 📋 Veri yapıları ve enum'lar
├── utils/                 # 🛠️ Yardımcı fonksiyonlar
├── workers/               # ☁️ Cloudflare Workers (TypeScript)
├── .github/workflows/     # 🤖 GitHub Actions CI/CD
└── main.go               # 🚀 Ana giriş noktası
```

### 📁 Katman Detayları

#### `/cmd` - 🎯 Giriş Noktaları
Projenin derlenebilir ana giriş noktalarını içerir.

#### `/configs` - ⚙️ Konfigürasyonlar
- `constants.go` - Proje sabitleri
- CORS, güvenlik başlıkları gibi statik ayarlar

#### `/database` - 🗄️ Veritabanı
- `init.go` - Veritabanı bağlantısı
- `/migrations` - Şema değişiklikleri

#### `/handlers` - 🌐 Sunum Katmanı
- HTTP isteklerini alır
- Veri doğrulaması yapar
- Servisleri çağırır
- JSON yanıtı döndürür

#### `/middlewares` - 🔗 Ara Katmanlar
Kimlik doğrulama, yetkilendirme, loglama gibi tüm istekleri etkileyen mantıklar

#### `/repositories` - 💾 Veri Erişim Katmanı
- Veritabanı iletişimini soyutlar
- SQL sorgularını içerir
- Servis katmanının veritabanı detaylarından habersiz kalmasını sağlar

#### `/services` - 🧠 İş Mantığı Katmanı
- Uygulamanın çekirdek mantığı
- Repository'leri kullanarak iş akışlarını tamamlar

#### `/types` - 📋 Veri Yapıları
- Struct'lar ve enum'lar
- API istek/yanıt modelleri
- Veritabanı tablolarının Go karşılıkları

#### `/utils` - 🛠️ Yardımcı Fonksiyonlar
- JWT oluşturma
- Şifre hash'leme
- Cookie yönetimi
- Katmandan bağımsız tekrar kullanılabilir kodlar

#### `/workers` - ☁️ Cloudflare Workers
- TypeScript tabanlı edge mantığı
- Statik varlık sunumu
- İstek filtreleme

#### `/.github/workflows` - 🤖 CI/CD
- `deploy.yml` - Otomatik deployment
- `daily-analytics-report.yml` - Zamanlanmış görevler

#### `main.go` - 🚀 Ana Dosya
- Bağımlılık enjeksiyonu
- Router ve middleware yapılandırması
- Web sunucusu başlatma

## 🚀 Kurulum ve Başlatma

### 📋 Ön Gereksinimler

```bash
# Gerekli araçları kurun
go install
docker install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### ⚙️ Yapılandırma

1. **Ortam değişkenlerini ayarlayın:**
   ```bash
   cp .env.example .env
   # .env dosyasını editleyerek değerleri kendinize göre düzenleyin
   ```

2. **Gerekli servisler:**
   ```bash
   # PostgreSQL ve Redis'i Docker ile başlatın
   docker-compose up -d postgres redis
   ```

### 🗄️ Veritabanı Migrasyonları

```bash
# Veritabanı şemasını oluştur/güncelle
go run cmd/migrate/main.go up

# Migrasyon durumunu kontrol et
go run cmd/migrate/main.go version

# Son migrasyonu geri al
go run cmd/migrate/main.go down 1
```

### 🚀 Sunucuyu Başlatma

```bash
# Bağımlılıkları kur
go mod tidy

# Sunucuyu başlat
go run main.go

# Veya derleyip çalıştır
go build -o server main.go
./server
```

### 🧪 Test

```bash
# Tüm testleri çalıştır
go test ./...

# Belirli bir paketi test et
go test ./services/...

# Test coverage
go test -cover ./...
```

### 📦 Production Build

```bash
# Optimized build
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server main.go

# Docker ile
docker build -t myapp .
docker run -p 8080:8080 myapp
```

## 🌍 Deployment

Proje, Hetzner VPS + Nginx + Cloudflare kombinasyonu için optimize edilmiştir:

1. **GitHub Actions** otomatik deployment'ı tetikler
2. **Hetzner VPS**'e kod deploy edilir
3. **Nginx** reverse proxy olarak çalışır
4. **Cloudflare** CDN ve güvenlik sağlar

## 📈 İzleme ve Loglama

- Structured logging (JSON format)
- Error tracking ve alerting
- Performance metrics
- Database query monitoring

## 🤝 Katkıda Bulunma

1. Fork'layın
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Değişikliklerinizi commit'leyin (`git commit -m 'Add amazing feature'`)
4. Branch'inizi push'layın (`git push origin feature/amazing-feature`)
5. Pull Request açın

## 📄 Lisans

Bu proje MIT lisansı altında lisanslanmıştır. Detaylar için `LICENSE` dosyasına bakın.

---

## 🔗 Faydalı Linkler

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/)
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/)
- [Cloudflare R2](https://developers.cloudflare.com/r2/)
- [JWT.io](https://jwt.io/)

## 💡 SSS (Sık Sorulan Sorular)

**S: Bu şablonu yeni bir proje için nasıl kullanırım?**
A: Repository'yi fork'layın veya template olarak kullanın, `.env` dosyasını yapılandırın ve `go run main.go` ile başlatın.

**S: Farklı bir veritabanı kullanabilir miyim?**
A: Evet, repository katmanını değiştirerek MySQL, MongoDB gibi alternatifler kullanabilirsiniz.

**S: Cloudflare yerine AWS S3 kullanabilir miyim?**
A: Evet, `utils/` klasöründeki dosya upload fonksiyonlarını S3 için yeniden yazabilirsiniz.
