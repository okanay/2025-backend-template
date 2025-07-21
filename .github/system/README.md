# 🚀 Ubuntu VPS - Nginx ve Systemd ile Backend Deployment Rehberi

Bu rehber, GoLang backend uygulamalarını Ubuntu VPS'e deploy etmek için gereken tüm konfigürasyonları ve komutları detaylarıyla açıklar.

## 📋 İçindekiler
1. [Nginx Konfigürasyonu](#nginx-konfigürasyonu)
2. [Systemd Servis Ayarları](#systemd-servis-ayarları)
3. [SSL Sertifikası](#ssl-sertifikası)
4. [Log İzleme](#log-izleme)
5. [Deployment Sırası](#deployment-sırası)
6. [Sorun Giderme](#sorun-giderme)

---

## 🔧 1. Nginx Konfigürasyonu

### Basit Template (Tek Backend)
**Dosya:** `/etc/nginx/sites-available/backend-template`

```nginx
server {
    listen 80;
    server_name template.hoi.com.tr;

    # Global ayarlar
    client_max_body_size 50m;  # Dosya yükleme limiti

    location / {
        proxy_pass http://localhost:4040;
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

    # Log dosyaları
    access_log /var/log/nginx/template-hoi-access.log;
    error_log /var/log/nginx/template-hoi-error.log;
}
```

### Gelişmiş Template (Çoklu Endpoint + Özellikler)

```nginx
# Load balancing için upstream tanımı
upstream backend_servers {
    least_conn;  # En az bağlantı olan sunucuya yönlendir
    server localhost:4040 weight=3;  # Ana sunucu (daha fazla ağırlık)
    server localhost:4041 weight=2;  # İkinci sunucu
    server localhost:4042 backup;    # Yedek sunucu (diğerleri çökerse)
}

server {
    listen 80;
    server_name template.hoi.com.tr;

    # Global ayarlar
    client_max_body_size 50m;

    # API endpoint'leri (daha katı ayarlar)
    location /api/ {
        # Load balancer kullan
        proxy_pass http://backend_servers;
        proxy_http_version 1.1;

        # Temel header'lar
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # API için özel header'lar
        proxy_set_header Content-Type application/json;
        proxy_set_header X-API-Version "v1";

        # Cloudflare
        proxy_set_header CF-Connecting-IP $http_cf_connecting_ip;
        proxy_set_header CF-IPCountry $http_cf_ipcountry;

        # API için kısa timeout (hızlı response beklentisi)
        proxy_connect_timeout 10s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # Admin paneli (IP bazlı erişim kontrolü)
    location /admin/ {
        # Sadece bu IP'lerden erişim izni
        allow 192.168.1.0/24;     # Yerel ağ
        allow 78.186.0.0/16;      # Türkiye IP aralığı örneği
        allow 95.70.162.47;       # Belirli IP (ofis IP'si)
        deny all;                 # Diğer tüm IP'leri reddet

        proxy_pass http://backend_servers;
        proxy_http_version 1.1;

        # Temel header'lar
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Admin için uzun timeout
        proxy_connect_timeout 60s;
        proxy_send_timeout 120s;
        proxy_read_timeout 120s;
    }

    # Dosya yükleme endpoint'i (büyük dosyalar)
    location /upload/ {
        client_max_body_size 100m;  # 100MB'a kadar dosya

        proxy_pass http://backend_servers;
        proxy_http_version 1.1;

        # Temel header'lar
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

        # Upload için uzun timeout
        proxy_connect_timeout 60s;
        proxy_send_timeout 300s;   # 5 dakika
        proxy_read_timeout 300s;   # 5 dakika
    }

    # WebSocket endpoint'i
    location /ws {
        proxy_pass http://backend_servers;
        proxy_http_version 1.1;

        # WebSocket için gerekli header'lar
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

        # WebSocket için çok uzun timeout (24 saat)
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;

        proxy_cache_bypass $http_upgrade;
    }

    # Static dosyalar (cache ile)
    location /static/ {
        proxy_pass http://backend_servers;

        # Cache ayarları
        expires 30d;                              # 30 gün cache
        add_header Cache-Control "public, immutable";
        add_header X-Cache-Status "HIT";

        # Static dosyalar için basit header'lar
        proxy_set_header Host $host;

        # Kısa timeout (static dosyalar hızlı olmalı)
        proxy_connect_timeout 10s;
        proxy_read_timeout 30s;
    }

    # Ana sayfa ve diğer tüm istekler
    location / {
        proxy_pass http://backend_servers;
        proxy_http_version 1.1;

        # WebSocket desteği (gerekirse)
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_cache_bypass $http_upgrade;

        # Temel header'lar
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Cloudflare header'ları
        proxy_set_header CF-Connecting-IP $http_cf_connecting_ip;
        proxy_set_header CF-Ray $http_cf_ray;
        proxy_set_header CF-IPCountry $http_cf_ipcountry;

        # Standart timeout
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Hata sayfaları
    error_page 502 503 504 /50x.html;
    location = /50x.html {
        root /var/www/html;
    }

    # Log dosyaları
    access_log /var/log/nginx/template-hoi-access.log;
    error_log /var/log/nginx/template-hoi-error.log;
}
```

### Rate Limiting (Dikkatli Kullanım)
**⚠️ Uyarı:** Rate limiting Nginx katmanında karmaşık olabilir. Basit senaryolar için:

```nginx
# nginx.conf içine ekle (http bloku içinde)
http {
    # IP başına saniyede 10 istek (10MB memory pool)
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
}

# site konfigürasyonunda kullan
location /api/ {
    # Burst: 20 isteğe kadar ani artışa izin ver
    # nodelay: Hemen işle, kuyruğa koyma
    limit_req zone=api burst=20 nodelay;

    # ... diğer proxy ayarları
}
```

**Rate limiting alternatifleri:**
- Backend uygulamada middleware ile yönet (önerilen)
- Cloudflare Rate Limiting kullan (daha kolay)
- Redis ile custom rate limiting

---

## ⚙️ 2. Systemd Servis Ayarları

### Basit Servis Template
**Dosya:** `/etc/systemd/system/backend-template.service`

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

# Log yönetimi
StandardOutput=journal
StandardError=journal

# GoLang graceful shutdown için
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30

[Install]
WantedBy=multi-user.target
```

### Çoklu Instance Servis (Load Balancing için)

**Port 4040 için servis:**
```ini
[Unit]
Description=Backend Template Go Service - Port 4040
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=backend
WorkingDirectory=/opt/backend-template
ExecStart=/opt/backend-template/main
Restart=always
RestartSec=10

# Environment ayarları
Environment=PORT=4040
Environment=INSTANCE_ID=1
EnvironmentFile=/opt/backend-template/.env

# Resource limitleri
LimitNOFILE=65536
LimitNPROC=4096

# Güvenlik
NoNewPrivileges=true
PrivateTmp=true

StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

**Port 4041 için servis:**
```ini
[Unit]
Description=Backend Template Go Service - Port 4041
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=backend
WorkingDirectory=/opt/backend-template
ExecStart=/opt/backend-template/main
Restart=always
RestartSec=10

Environment=PORT=4041
Environment=INSTANCE_ID=2
EnvironmentFile=/opt/backend-template/.env

LimitNOFILE=65536
LimitNPROC=4096

NoNewPrivileges=true
PrivateTmp=true

StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

---

## 🔒 3. SSL Sertifikası (Let's Encrypt)

```bash
# 1. Sistem güncelle
sudo apt update

# 2. Certbot kur
sudo apt install certbot python3-certbot-nginx -y

# 3. SSL sertifikası al (otomatik nginx konfigürasyonu)
sudo certbot --nginx -d template.hoi.com.tr

# 4. Otomatik yenileme test et
sudo certbot renew --dry-run
```

**SSL sonrası Nginx otomatik eklemeleri:**
```nginx
listen 443 ssl;
ssl_certificate /etc/letsencrypt/live/template.hoi.com.tr/fullchain.pem;
ssl_certificate_key /etc/letsencrypt/live/template.hoi.com.tr/privkey.pem;
```

---

## 📊 4. Log İzleme ve Debugging

### Nginx Log Komutları
```bash
# Canlı erişim logları
sudo tail -f /var/log/nginx/template-hoi-access.log

# Hata logları
sudo tail -f /var/log/nginx/template-hoi-error.log

# Son 100 satır
sudo tail -n 100 /var/log/nginx/template-hoi-access.log

# Belirli IP'nin istekleri
grep "95.70.162.47" /var/log/nginx/template-hoi-access.log

# 404 hataları
grep " 404 " /var/log/nginx/template-hoi-access.log

# En çok istek yapan IP'ler (top 10)
awk '{print $1}' /var/log/nginx/template-hoi-access.log | sort | uniq -c | sort -nr | head -10
```

### Systemd Service Log Komutları
```bash
# Canlı servis logları
sudo journalctl -u backend-template -f

# Son 100 satır
sudo journalctl -u backend-template -n 100

# Sadece bugünün logları
sudo journalctl -u backend-template --since today

# Belirli zaman aralığı
sudo journalctl -u backend-template --since "2025-01-15 10:00" --until "2025-01-15 11:00"

# Hata içeren loglar
sudo journalctl -u backend-template --grep="error"

# Tüm servisler için özet durum
sudo systemctl status
```

---

## 🚀 5. Deployment Sırası

### Kurulum Adımları

```bash
# 1. NGINX KURULUMU
sudo apt update
sudo apt install nginx -y

# 2. NGINX KONFİGÜRASYON OLUŞTUR
sudo nano /etc/nginx/sites-available/backend-template

# 3. KONFİGÜRASYONU AKTİFLEŞTİR
sudo ln -s /etc/nginx/sites-available/backend-template /etc/nginx/sites-enabled/

# 4. NGINX KONFİGÜRASYONUNU TEST ET
sudo nginx -t

# 5. NGINX'İ BAŞLAT
sudo systemctl restart nginx
sudo systemctl enable nginx

# 6. BACKEND UYGULAMASI HAZIRLA
# (Binary dosyasını sunucuya yükle)
# .env dosyasını hazırla

# 7. SYSTEMd SERVİS OLUŞTUR
sudo nano /etc/systemd/system/backend-template.service

# 8. SERVİS AYARLARINI YÜKLE
sudo systemctl daemon-reload

# 9. SERVİSİ BAŞLAT
sudo systemctl start backend-template
sudo systemctl enable backend-template

# 10. SSL SERTİFİKASI KUR
sudo certbot --nginx -d template.hoi.com.tr

# 11. DURUM KONTROL ET
sudo systemctl status backend-template
sudo systemctl status nginx
```

### Çoklu Instance Deployment (Load Balancing)
```bash
# Her instance için ayrı servis dosyası oluştur
sudo nano /etc/systemd/system/backend-template-4040.service
sudo nano /etc/systemd/system/backend-template-4041.service
sudo nano /etc/systemd/system/backend-template-4042.service

# Tüm servisleri yükle
sudo systemctl daemon-reload

# Servisleri başlat
sudo systemctl start backend-template-4040
sudo systemctl start backend-template-4041
sudo systemctl start backend-template-4042

# Otomatik başlatmayı etkinleştir
sudo systemctl enable backend-template-4040
sudo systemctl enable backend-template-4041
sudo systemctl enable backend-template-4042

# Durumları kontrol et
sudo systemctl status backend-template-*
```

---

## 🚨 6. Sorun Giderme

### Nginx Sorunları
```bash
# Konfigürasyon hatası kontrolü
sudo nginx -t

# Nginx durumu
sudo systemctl status nginx

# Nginx yeniden başlatma
sudo systemctl restart nginx

# Port kullanım kontrolü
sudo netstat -tlnp | grep :80
sudo netstat -tlnp | grep :443

# Nginx process kontrolü
ps aux | grep nginx
```

### Backend Service Sorunları
```bash
# Detaylı hata mesajı
sudo journalctl -u backend-template --no-pager -l

# Servis durumu
sudo systemctl status backend-template

# Servis yeniden başlatma
sudo systemctl restart backend-template

# Port kullanımda mı?
sudo netstat -tlnp | grep :4040

# Process kontrol
ps aux | grep main
```

### Load Balancing Sorunları
```bash
# Hangi instance'lar çalışıyor
sudo systemctl status backend-template-*

# Port kontrolleri
sudo netstat -tlnp | grep -E ":(4040|4041|4042)"

# Upstream durumu test (manuel)
curl -I http://localhost:4040/health
curl -I http://localhost:4041/health
curl -I http://localhost:4042/health
```

### SSL Sorunları
```bash
# SSL sertifika durumu
sudo certbot certificates

# SSL test
curl -I https://template.hoi.com.tr

# SSL yenileme test
sudo certbot renew --dry-run

# SSL logları
sudo tail -f /var/log/letsencrypt/letsencrypt.log
```

### Yaygın Hatalar ve Çözümler

**502 Bad Gateway:**
```bash
# Backend çalışıyor mu?
sudo systemctl status backend-template

# Port dinleniyor mu?
sudo netstat -tlnp | grep :4040

# Firewall engelliyor mu?
sudo ufw status
```

**403 Forbidden (Admin sayfasında):**
```bash
# IP whitelist kontrol et
# Nginx konfigürasyonunda allow/deny satırlarını gözden geçir

# Client IP'sini log'dan öğren
sudo tail -f /var/log/nginx/template-hoi-access.log
```

**SSL sertifikası alınamıyor:**
```bash
# Domain DNS ayarları doğru mu?
nslookup template.hoi.com.tr

# Port 80 ve 443 açık mı?
sudo ufw status
sudo netstat -tlnp | grep -E ":(80|443)"
```

---

## 🎯 Hızlı Referans Komutları

```bash
# Nginx
sudo nginx -t                    # Konfigürasyon test
sudo systemctl restart nginx    # Nginx yeniden başlat
sudo tail -f /var/log/nginx/template-hoi-access.log

# Backend Service
sudo systemctl restart backend-template
sudo journalctl -u backend-template -f
sudo systemctl status backend-template

# SSL
sudo certbot renew --dry-run
sudo certbot certificates

# Network
sudo netstat -tlnp | grep :4040
ps aux | grep main
```

Bu rehber sayesinde Ubuntu VPS'inizde profesyonel bir şekilde backend deployment yapabilir, load balancing, IP kısıtlaması ve cache optimizasyonları uygulayabilirsiniz.
