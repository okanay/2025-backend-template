# 🚀 GitHub Actions ile Sıfır Kesintili (Zero-Downtime) Deployment

Bu rehber, GoLang backend uygulamalarını Ubuntu VPS'e systemd kullanmadan, sıfır kesinti ile deploy etmek için gereken tüm sunucu yapılandırmasını ve GitHub Actions iş akışını açıklar. Bu yöntem, "Blue-Green Deployment" stratejisinin basitleştirilmiş bir versiyonunu kullanır.

## 📋 İçindekiler

1. [Konsept: Sıfır Kesintili Deployment Nasıl Çalışır?](#konsept)
2. [Gerekli Sunucu Yapılandırması](#sunucu-yapılandırması)
3. [GitHub Actions Workflow (deploy.yml)](#github-actions-workflow)
4. [Manuel Kontrol ve Sorun Giderme](#manuel-kontrol)
5. [İlk Deploy Adımları](#ilk-deploy-adımları)

## 💡 1. Konsept: Sıfır Kesintili Deployment Nasıl Çalışır? {#konsept}

Bu sistem, systemd gibi statik servisler yerine, uygulama süreçlerini doğrudan script ile yöneterek kesintisiz geçiş yapmayı hedefler. Deploy.yml dosyasında tanımlı **PORT_A (4040)** ve **PORT_B (4041)** arasında geçiş yapmaktır.

### Olay Akışı:

1. **Mevcut Durum**: Canlı uygulamanız Port A'da (4040) çalışıyor ve Nginx tüm trafiği bu porta yönlendiriyor.

2. **Deploy Başlar**: GitHub Actions workflow'u manuel olarak tetiklenir (veya main branch'ine kod gönderilirse).

3. **Yeni Versiyon Hazırlanır**: Script, sunucuda yeni versiyonu kurar ve boşta olan Port B'de (4041) başlatır.

4. **Sağlık Kontrolü**: Script, Port B'deki yeni uygulamanın `/` endpoint'ini curl ile test eder.

5. **Anlık Geçiş (Soft Switch)**: Sağlık kontrolü başarılı olursa, script Nginx'in yapılandırmasını güncelleyerek gelen tüm yeni trafiği anında Port B'ye yönlendirir. Bu işlem `nginx -s reload` komutu sayesinde mevcut bağlantıları kesmeden yapılır.

6. **Eski Versiyon Durdurulur**: Geçiş tamamlandıktan sonra, artık trafik almayan Port A'daki eski uygulama güvenli bir şekilde sonlandırılır.

7. **Temizlik**: `KEEP_VERSIONS: 3` ayarına göre, 3'ten eski release klasörleri sunucudan silinir.

Bir sonraki deploy'da bu süreç tersine işler: Canlı uygulama Port B'de çalışırken, yeni versiyon Port A'da test edilir ve geçiş yapılır.

## 🔧 2. Gerekli Sunucu Yapılandırması {#sunucu-yapılandırması}

> ⚠️ **ÖNEMLI**: Aşağıdaki yapılandırmalar deploy.yml dosyasındaki ayarlarla tam uyumlu olmalıdır.

### ⚙️ Deploy.yml ile Senkronize Edilmesi Gereken Ayarlar

Bu ayarlar deploy.yml dosyanızdaki `env:` bloğu ile **birebir eş olmalıdır**:

```yaml
env:
  PROJECT_ROOT: "/root/2025-backend-template"  # ← Bu yolu sunucuda oluşturacaksınız
  VERSIONS_DIR_NAME: "build-versions"         # ← Otomatik oluşturulur
  STATE_DIR_NAME: "build-state"               # ← Otomatik oluşturulur
  PORT_A: 4040                                # ← Nginx konfigürasyonunda kullanılacak
  PORT_B: 4041                                # ← Nginx konfigürasyonunda kullanılacak
  KEEP_VERSIONS: 3                            # ← Temizlik ayarı
  HEALTH_CHECK_PATH: "/"                      # ← Sağlık kontrolü endpoint'i
```

### Nginx Kurulumu ve Dinamik Yönlendirme Ayarları

#### Adım 1: Gerekli Dizini Oluşturun

```bash
sudo mkdir -p /etc/nginx/includes
```

#### Adım 2: Ana Nginx Konfigürasyonunu Oluşturun

Dosya: `/etc/nginx/sites-available/backend-template`

> 🔧 **SİZİN DEĞİŞTİRMENİZ GEREKEN ALANLAR:**
> - `server_name` satırını kendi domain'iniz ile değiştirin
> - Log dosyası yollarını domain'inize göre güncelleyin

```nginx
# 🔧 BURAYI DEĞİŞTİRİN: Her proje için benzersiz upstream adı
# Örnek: api_backend, admin_backend, user_backend vb.
upstream PROJE_ADI_backend {
    # Port bilgisi, deploy script'i tarafından güncellenecek bu dosyadan okunacak.
    # 🔧 BURAYI DEĞİŞTİRİN: Her proje için farklı dosya adı
    include /etc/nginx/includes/PROJE_ADI_upstream.conf;
}

server {
    listen 80;
    # SSL aktif edildiğinde bu satır otomatik olarak 443'e güncellenir.

    # 🔧 BURAYI DEĞİŞTİRİN: Kendi domain adınızı yazın
    server_name SIZIN-DOMAIN-ADINIZ.com;

    location / {
        # 🔧 BURAYI DEĞİŞTİRİN: Yukarıdaki upstream adı ile eş olmalı
        proxy_pass http://PROJE_ADI_backend;
        proxy_http_version 1.1;

        # WebSocket desteği
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_cache_bypass $http_upgrade;

        # Temel header'lar
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Cloudflare header'ları (CDN kullanıyorsanız)
        proxy_set_header CF-Connecting-IP $http_cf_connecting_ip;
        proxy_set_header CF-Ray $http_cf_ray;
        proxy_set_header CF-IPCountry $http_cf_ipcountry;

        # GoLang için timeout ayarları
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # 🔧 BURAYI DEĞİŞTİRİN: Log dosyası yollarını domain'inize göre güncelleyin
    access_log /var/log/nginx/SIZIN-DOMAIN-ADINIZ-access.log;
    error_log /var/log/nginx/SIZIN-DOMAIN-ADINIZ-error.log;
}
```

#### Adım 3: Başlangıç Port Dosyasını Oluşturun

Deploy script'i bu dosyanın içeriğini deploy.yml'deki `PORT_A` ve `PORT_B` değerleri arasında geçiş yapar.

```bash
# 🔧 BURAYI DEĞİŞTİRİN: Nginx konfigürasyonundaki dosya adı ile aynı olmalı
# Deploy.yml'deki PORT_A değeri (4040) ile başlangıç yapın
echo "server 127.0.0.1:4040;" | sudo tee /etc/nginx/includes/PROJE_ADI_upstream.conf
```

> ⚠️ **ÇOK ÖNEMLİ**: Aynı sunucuda birden fazla proje varsa:
> - Her proje için farklı upstream adı kullanın (`api_backend`, `admin_backend` vb.)
> - Her proje için farklı upstream dosyası oluşturun (`api_upstream.conf`, `admin_upstream.conf`)
> - Her proje için farklı port aralıkları kullanın (Proje1: 4040-4041, Proje2: 5050-5051)

#### Adım 4: Nginx'i Aktifleştirin ve Test Edin

```bash
# Oluşturduğunuz konfigürasyonu aktifleştirin
sudo ln -s /etc/nginx/sites-available/backend-template /etc/nginx/sites-enabled/

# Yazdığınız konfigürasyonda bir hata olup olmadığını kontrol edin
sudo nginx -t

# Her şey yolundaysa Nginx'i yeniden başlatın
sudo systemctl restart nginx
```

### Proje ve Durum Klasörleri

#### 1. Proje Ana Dizini (PROJECT_ROOT)

Deploy.yml dosyasındaki `PROJECT_ROOT: "/root/2025-backend-template"` değeri ile eş olmalıdır.

```bash
# Deploy.yml'deki PROJECT_ROOT değeri
mkdir -p /root/2025-backend-template
```

> 🔧 **SİZİN DEĞİŞTİRMENİZ GEREKEN ALAN:**
> Deploy.yml dosyasındaki `PROJECT_ROOT` değerini kendi sunucu yolunuza göre güncelleyin.

#### 2. Merkezi .env Dosyası

Tüm hassas bilgileri içeren `.env` dosyanızı deploy.yml'deki `PROJECT_ROOT` konumuna yerleştirin:

```bash
# Örnek konum (deploy.yml'deki PROJECT_ROOT'a göre)
/root/2025-backend-template/.env
```

Bu dosya Git'e dahil edilmemelidir. Deploy script'i her versiyonu kurarken bu merkezi dosyayı kopyalar.

#### 3. Otomatik Oluşturulan Klasörler

Deploy script'i ilk çalıştığında deploy.yml'deki ayarlara göre otomatik olarak oluşturur:

- **build-versions** (`VERSIONS_DIR_NAME`): Her deploy'un zaman damgalı kopyaları
- **build-state** (`STATE_DIR_NAME`): Sistem durumu (`live.port` ve `app.pid` dosyaları)

## ⚙️ 3. GitHub Actions Workflow (deploy.yml) {#github-actions-workflow}

### 🔧 SİZİN YAPMANIZ GEREKEN AYARLAR

Deploy.yml dosyasında aşağıdaki değerleri **kendi projenize göre güncelleyin**:

```yaml
env:
  # 🔧 BURAYІ DEĞİŞTİRİN: Sunucudaki proje yolunuz
  PROJECT_ROOT: "/root/2025-backend-template"

  # 🔧 İSTERSENİZ DEĞİŞTİRİN: Klasör isimleri
  VERSIONS_DIR_NAME: "build-versions"
  STATE_DIR_NAME: "build-state"

  # 🔧 İSTERSENİZ DEĞİŞTİRİN: Port numaraları (aynı sunucuda farklı projeler için farklı aralıklar)
  PORT_A: 4040
  PORT_B: 4041

  # 🔧 İSTERSENİZ DEĞİŞTİRİN: Tutulacak versiyon sayısı
  KEEP_VERSIONS: 3

  # 🔧 BURAYІ DEĞİŞTİRİN: GoLang app'inizin sağlık kontrolü endpoint'i
  HEALTH_CHECK_PATH: "/"

  # 🔧 BURAYІ DEĞİŞTİRİN: Projenizin kurulum ve build komutları
  INSTALL_COMMAND: "go mod tidy"
  BUILD_COMMAND: "/usr/local/go/bin/go build -o main ."

  # 🔧 BURAYІ DEĞİŞTİRİN: Nginx upstream dosya adı (Nginx konfigürasyonu ile eş olmalı)
  UPSTREAM_CONF_FILE: "/etc/nginx/includes/PROJE_ADI_upstream.conf"
```

### GitHub Secrets Ayarları

GitHub projenizin **Settings > Secrets and variables > Actions** menüsünden şu secrets'ları tanımlayın:

- `HOST`: Sunucu IP adresi
- `USERNAME`: SSH kullanıcı adı (genellikle root)
- `PRIVATE_KEY`: SSH private key
- `PASSPHRASE`: SSH key parolası (varsa)
- `PORT`: SSH portu (genellikle 22)

### Workflow İşleyişi

Deploy.yml çalıştığında loglarda şu adımları göreceksiniz:

1. **Ortamı Hazırla**: `PROJECT_ROOT`, `VERSIONS_DIR_NAME`, `STATE_DIR_NAME` klasörlerini oluştur
2. **Kodu Klonla**: GitHub repository'yi yeni release klasörüne kopyala
3. **.env Kopyala**: Merkezi .env dosyasını yeni versiyona kopyala
4. **Bağımlılıkları Yükle**: `INSTALL_COMMAND` komutunu çalıştır
5. **Build Et**: `BUILD_COMMAND` ile uygulamayı derle
6. **Port Belirle**: `build-state/live.port` dosyasından mevcut portu oku, diğerini seç
7. **Uygulamayı Başlat**: Yeni portta uygulamayı çalıştır
8. **Sağlık Kontrolü**: `HEALTH_CHECK_PATH` endpoint'ini test et
9. **Nginx Geçişi**: Backend upstream'i yeni porta yönlendir
10. **Eski Süreci Durdur**: Önceki port'taki uygulamayı sonlandır
11. **Durumu Güncelle**: `build-state` dosyalarını güncelle
12. **Temizlik**: `KEEP_VERSIONS` sayısına göre eski versiyonları sil

## 🔍 4. Manuel Kontrol ve Sorun Giderme {#manuel-kontrol}

> ⚠️ **Not**: Aşağıdaki komutlarda yollar deploy.yml'deki `PROJECT_ROOT` değerinize göre güncellenmelidir.

### Hangi port canlıda?
```bash
# Deploy.yml'deki PROJECT_ROOT + STATE_DIR_NAME yolu
cat /root/2025-backend-template/build-state/live.port
```

### Çalışan uygulamanın PID'si nedir?
```bash
cat /root/2025-backend-template/build-state/app.pid
```

### Deploy.yml'deki portları acil durdurmak için:
```bash
# PORT_A ve PORT_B değerlerinize göre
sudo fuser -k 4040/tcp
sudo fuser -k 4041/tcp
```

### Son deploy'un uygulama loglarını görmek için:
```bash
# Deploy.yml'deki PROJECT_ROOT + VERSIONS_DIR_NAME yolu
LATEST_RELEASE=$(ls -1tr /root/2025-backend-template/build-versions | tail -n 1)
cat /root/2025-backend-template/build-versions/$LATEST_RELEASE/app.log
```

### Nginx upstream durumunu kontrol edin:
```bash
# 🔧 BURAYI DEĞİŞTİRİN: Kendi upstream dosya adınızı yazın
cat /etc/nginx/includes/PROJE_ADI_upstream.conf
# Çıktı: server 127.0.0.1:4040; veya server 127.0.0.1:4041;
```

### Aynı sunucuda birden fazla proje durumu:
```bash
# Tüm upstream dosyalarını listele
ls -la /etc/nginx/includes/

# Örnek çıktı:
# api_upstream.conf     -> server 127.0.0.1:4040;
# admin_upstream.conf   -> server 127.0.0.1:5050;
# user_upstream.conf    -> server 127.0.0.1:6060;
```

## ✅ 5. İlk Deploy Adımları {#ilk-deploy-adımları}

### Sunucu Hazırlığı
1. Ubuntu sunucunuza Nginx'i kurun: `sudo apt install nginx`
2. Yukarıdaki **Nginx konfigürasyonunu** yapın ve şunları güncelleyin:
   - Domain adınızı güncelleyin
   - Upstream adını projenize özel yapın (`PROJE_ADI_backend`)
   - Upstream dosya adını güncelleyin (`PROJE_ADI_upstream.conf`)
3. Deploy.yml'deki `PROJECT_ROOT` klasörünü oluşturun
4. Merkezi `.env` dosyanızını PROJECT_ROOT'a yerleştirin

### GitHub Ayarları
5. GitHub Secrets'ları tanımlayın (`HOST`, `USERNAME`, `PRIVATE_KEY`, vb.)
6. Deploy.yml dosyasındaki tüm `🔧` işaretli alanları kendi projenize göre güncelleyin
7. Özellikle şunları kontrol edin:
   - `PROJECT_ROOT`: Sunucudaki proje yolu
   - `UPSTREAM_CONF_FILE`: Nginx upstream dosya yolu
   - `PORT_A` ve `PORT_B`: Diğer projelerle çakışmayan portlar
   - `HEALTH_CHECK_PATH`: GoLang app'inizin test endpoint'i
   - `BUILD_COMMAND`: Doğru build komutu

### Çoklu Proje Örneği
Aynı sunucuda birden fazla proje çalıştırıyorsanız:

**Proje 1 (API Backend):**
- Upstream: `api_backend`
- Portlar: 4040-4041
- Dosya: `/etc/nginx/includes/api_upstream.conf`

**Proje 2 (Admin Panel):**
- Upstream: `admin_backend`
- Portlar: 5050-5051
- Dosya: `/etc/nginx/includes/admin_upstream.conf`

**Proje 3 (User Service):**
- Upstream: `user_backend`
- Portlar: 6060-6061
- Dosya: `/etc/nginx/includes/user_upstream.conf`

### Deploy Testi
8. GitHub Actions sekmesinden "Deploy to Production with Zero Downtime" workflow'unu manuel olarak çalıştırın
9. Deploy loglarını canlı takip edin
10. Deploy başarılı olduktan sonra domain'inizi browser'da test edin

### Sorun Giderme
Herhangi bir sorun yaşarsanız:
- Deploy loglarını kontrol edin
- Manuel kontrol komutlarını kullanın
- Nginx error loglarını inceleyin: `sudo tail -f /var/log/nginx/SIZIN-DOMAIN-ADINIZ-error.log`
