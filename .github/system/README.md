# Ubuntu Backend Template Deployment Rehberi

Bu rehber, GoLang backend template'ini Ubuntu sunucusuna deploy etmek için gereken tüm adımları detaylarıyla açıklar.

## 📁 Klasör Yapısı
```
system/
├── README.md (bu dosya)
├── nginx/
│   └── backend-template.conf
├── systemd/
│   └── backend-template.service
└── ssl/
    └── certbot-setup.md
```

## 🔧 1. Nginx Konfigürasyonu

### Dosya: `/etc/nginx/sites-available/backend-template`

Nginx, web sunucusu olarak HTTP isteklerini alıp backend uygulamanıza proxy eder.

**Konfigürasyon detayları:**
- **Port 80**: HTTP trafiğini dinler
- **server_name**: Domain adınız (template.hoi.com.tr)
- **proxy_pass**: Backend uygulamanızın çalıştığı port (localhost:4040)
- **Cloudflare Headers**: CDN üzerinden gelen gerçek IP ve ülke bilgilerini backend'e iletir
- **Timeout Ayarları**: GoLang uygulamaları için optimize edilmiş zaman aşımı değerleri
- **Log Dosyaları**: Erişim ve hata loglarını ayrı dosyalarda saklar

### Template - Nginx Konfigürasyonu:

```nginx
server {
    listen 80;
    server_name template.hoi.com.tr;

    location / {
        proxy_pass http://localhost:4040;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;

        # CLOUDFLARE HEADER'LARI
        proxy_set_header CF-Connecting-IP $http_cf_connecting_ip;
        proxy_set_header CF-Ray $http_cf_ray;
        proxy_set_header CF-IPCountry $http_cf_ipcountry;
        proxy_set_header CF-Visitor $http_cf_visitor;
        proxy_set_header True-Client-IP $http_true_client_ip;

        # Timeout ayarları (GoLang uygulamaları için önemli)
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    access_log /var/log/nginx/template-hoi-access.log;
    error_log /var/log/nginx/template-hoi-error.log;
}
```

### Kurulum Adımları:

```bash
# 1. Nginx konfigürasyon dosyasını oluştur
sudo nano /etc/nginx/sites-available/backend-template

# 2. Konfigürasyonu sites-enabled'a linkle
sudo ln -s /etc/nginx/sites-available/backend-template /etc/nginx/sites-enabled/

# 3. Nginx konfigürasyonunu test et
sudo nginx -t

# 4. Nginx'i yeniden başlat
sudo systemctl restart nginx
```

**Nginx test komutu (-t flag):**
- Konfigürasyon dosyalarında syntax hatası var mı kontrol eder
- Hata varsa satır numarasını gösterir
- "test is successful" mesajı görürseniz konfigürasyon doğru

## ⚙️ 2. Systemd Servis Konfigürasyonu

### Dosya: `/etc/systemd/system/backend-template.service`

Systemd, Linux'ta servisleri yöneten sistem. Backend uygulamanızı otomatik başlatır, çökerse yeniden başlatır.

**Servis bölümleri açıklaması:**

**[Unit] Bölümü:**
- `Description`: Servis açıklaması
- `After`: Hangi servislerden sonra başlayacağını belirtir (network ve postgresql)
- `Wants`: Bağımlı servisler (postgresql gerekli ama zorunlu değil)

**[Service] Bölümü:**
- `Type=simple`: Uygulama direkt çalışır, fork etmez
- `User=root`: Hangi kullanıcıyla çalışacağı
- `WorkingDirectory`: Uygulamanın çalışacağı dizin
- `ExecStart`: Başlatılacak komut
- `Restart=on-failure`: Sadece hatayla kapandığında yeniden başlat
- `RestartSec=5`: Yeniden başlatma öncesi 5 saniye bekle
- `EnvironmentFile`: .env dosyasını yükle
- `KillMode=mixed`: Process'i nasıl sonlandıracağı
- `KillSignal=SIGTERM`: Graceful shutdown için sinyal
- `TimeoutStopSec=30`: Zorla kapatmadan önce 30 saniye bekle

**[Install] Bölümü:**
- `WantedBy=multi-user.target`: Sistem boot edildiğinde otomatik başlat

### Template - Systemd Service:

```ini
[Unit]
Description=Backend Template Go Service
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/root/2025-backend-template
ExecStart=/root/2025-backend-template/main
Restart=on-failure
RestartSec=5
EnvironmentFile=/root/2025-backend-template/.env
StandardOutput=journal
StandardError=journal

# GoLang uygulamaları için önemli ayarlar
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30

[Install]
WantedBy=multi-user.target
```

### Kurulum Adımları:

```bash
# 1. Servis dosyasını oluştur
sudo nano /etc/systemd/system/backend-template.service

# 2. Systemd'ye yeni servis dosyasını tanıt
sudo systemctl daemon-reload

# 3. Servisi sistem başlangıcında otomatik başlat
sudo systemctl enable backend-template

# 4. Servisi başlat
sudo systemctl start backend-template

# 5. Servis durumunu kontrol et
sudo systemctl status backend-template
```

**Servis komutları ve durumları:**

- `start`: Servisi başlatır
- `stop`: Servisi durdurur
- `restart`: Servisi yeniden başlatır
- `reload`: Konfigürasyonu yeniden yükler (uygulamaya bağlı)
- `status`: Servis durumunu gösterir
- `enable`: Boot'ta otomatik başlatmayı etkinleştirir
- `disable`: Boot'ta otomatik başlatmayı devre dışı bırakır

**Status çıktısındaki durumlar:**
- `active (running)`: Çalışıyor
- `inactive (dead)`: Durdurulmuş
- `failed`: Hata nedeniyle durmuş
- `activating`: Başlatılıyor

## 🔒 3. SSL Sertifikası Kurulumu

### Let's Encrypt ile Ücretsiz SSL

SSL sertifikası, HTTPS bağlantısı için gereklidir. Certbot, Let's Encrypt'ten otomatik sertifika alır ve yeniler.

```bash
# 1. Sistem paketlerini güncelle
sudo apt update

# 2. Certbot ve Nginx eklentisini kur
sudo apt install certbot python3-certbot-nginx

# 3. Domain için SSL sertifikası al
sudo certbot --nginx -d template.hoi.com.tr
```

**Certbot ne yapar:**
- Nginx konfigürasyonunu otomatik düzenler
- Port 443 (HTTPS) için SSL ayarları ekler
- HTTP'den HTTPS'e yönlendirme ekler
- Sertifika dosyalarını `/etc/letsencrypt/` altında saklar
- Otomatik yenileme için cron job kurar

**SSL kurulumu sonrası Nginx konfigürasyonu otomatik olarak şu eklemeleri içerir:**
```nginx
listen 443 ssl;
ssl_certificate /etc/letsencrypt/live/domain/fullchain.pem;
ssl_certificate_key /etc/letsencrypt/live/domain/privkey.pem;
```

## 📊 4. Log İzleme ve Debugging

### Nginx Logları:
```bash
# Erişim loglarını takip et
sudo tail -f /var/log/nginx/template-hoi-access.log

# Hata loglarını takip et
sudo tail -f /var/log/nginx/template-hoi-error.log
```

### Systemd Servis Logları:
```bash
# Servis loglarını görüntüle
sudo journalctl -u backend-template -f

# Son 100 satır log
sudo journalctl -u backend-template -n 100

# Belirli tarih aralığında loglar
sudo journalctl -u backend-template --since "2025-01-01" --until "2025-01-02"
```

## 🔄 5. Deployment Sırası

1. **Backend uygulamasını derle** (lokal makinede)
2. **Binary'yi sunucuya yükle**
3. **Environment dosyasını (.env) hazırla**
4. **Nginx konfigürasyonunu oluştur ve test et**
5. **Systemd servisini kur ve başlat**
6. **SSL sertifikası kur**
7. **Logları kontrol ederek test et**

## 🚨 Sık Karşılaşılan Sorunlar

### Backend servis başlamıyor:
```bash
# Detaylı hata mesajını gör
sudo journalctl -u backend-template --no-pager

# Port kullanımda mı kontrol et
sudo netstat -tlnp | grep :4040

# Firewall engelliyor mu?
sudo ufw status
```

### Nginx 502 Bad Gateway:
- Backend servisi çalışmıyor olabilir
- Port numarası yanlış olabilir
- Firewall backend portunu engelliyor olabilir

### SSL sertifikası alınamıyor:
- Domain DNS ayarları doğru mu?
- Port 80 ve 443 açık mı?
- Nginx'te aynı domain için çakışma var mı?

## 📝 Önemli Notlar

- **Port 4040**: Backend uygulamanızın dinlediği port
- **Domain**: template.hoi.com.tr yerine kendi domain'inizi kullanın
- **Cloudflare**: CDN kullanıyorsanız header'lar önemli
- **Environment**: .env dosyası servis dosyasında tanımlı
- **Logs**: Hata durumlarında ilk kontrol edilecek yer
- **Security**: Root kullanıcısı yerine dedicated user oluşturmak daha güvenli
