-- DOSYA TABLOLARI (Tüm dosyalar için genel tablo)
CREATE TABLE IF NOT EXISTS files (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    url TEXT NOT NULL, -- Erişim URL'i
    file_type TEXT NOT NULL, -- Dosya MIME tipi
    filename TEXT NOT NULL, -- Orijinal dosya adı
    file_category TEXT, -- 'cv', 'cover_letter', 'certificate', 'image', vb.
    size_in_bytes INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'active', 'deleted', 'orphaned'
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    UNIQUE (url)
);

-- TEK UPLOAD SIGNATURES TABLOSU (Tüm dosya yüklemeleri için)
CREATE TABLE IF NOT EXISTS files_signatures (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    presigned_url TEXT NOT NULL, -- Yükleme için ön-imzalı URL
    upload_url TEXT NOT NULL, -- Dosyanın erişileceği URL
    filename TEXT NOT NULL, -- Dosya adı
    file_type TEXT NOT NULL, -- MIME tipi
    file_category TEXT, -- 'cv', 'image', vb.
    expires_at TIMESTAMPTZ NOT NULL, -- İmzanın geçerlilik süresi
    completed BOOLEAN DEFAULT FALSE, -- Yükleme tamamlandı mı?
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL
);
