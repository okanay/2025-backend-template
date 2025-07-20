package AutomationService

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// JobFunc, zamanlandığında çalıştırılacak olan fonksiyonun tipidir.
type JobFunc func()

// jobState, bir işin mevcut durumunu belirtir.
type jobState string

const (
	StateIdle    jobState = "idle"
	StateRunning jobState = "running"
)

// jobInfo, tek bir işin tüm bilgilerini ve durumunu barındırır.
type jobInfo struct {
	Key         string
	IsRecurring bool
	Job         JobFunc
	EntryID     cron.EntryID
	State       jobState
}

// AutomationService, periyodik ve tek seferlik görevleri güvenli bir şekilde yönetir.
type AutomationService struct {
	mu       sync.RWMutex
	cron     *cron.Cron
	registry map[string]*jobInfo // Tüm işlerin merkezi kayıt defteri
}

// NewService, yeni bir otomasyon servisi oluşturur ve başlatır.
func NewService() *AutomationService {
	service := &AutomationService{
		cron:     cron.New(cron.WithSeconds()), // Saniye hassasiyetini aktif et
		registry: make(map[string]*jobInfo),
	}
	service.cron.Start()
	log.Println("✅ [AUTOMATION] Service started and running.")
	return service
}

// Stop, tüm zamanlanmış görevleri durdurur ve kaynakları temizler.
func (s *AutomationService) Stop() {
	s.cron.Stop()
	log.Println("🛑 [AUTOMATION] Service stopped.")
}

// Add, sisteme periyodik olarak çalışacak statik bir iş ekler.
// Genellikle sunucu başlangıcında çağrılır.
// Spec: Standart cron formatı (örn: "@every 6h", "0 4 * * *").
func (s *AutomationService) Add(key string, spec string, job JobFunc) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Wrapper fonksiyon, asıl işin etrafında durum kontrolü yapar.
	wrappedJob := s.createJobWrapper(key)

	entryID, err := s.cron.AddFunc(spec, wrappedJob)
	if err != nil {
		log.Printf("[AUTOMATION] ERROR adding cron job for key '%s': %v", key, err)
		return err
	}

	// İşi merkezi kayıt defterine ekle.
	s.registry[key] = &jobInfo{
		Key:         key,
		IsRecurring: true,
		Job:         job,
		EntryID:     entryID,
		State:       StateIdle,
	}

	log.Printf("[AUTOMATION] Added recurring job: '%s' with spec '%s'", key, spec)
	return nil
}

// Schedule, tek seferlik bir işi gelecekteki belirli bir zamanda çalışmak üzere zamanlar.
// Eğer aynı anahtarla mevcut bir zamanlama varsa, üzerine yazar.
func (s *AutomationService) Schedule(key string, at time.Time, job JobFunc) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Eğer bu key ile eski bir görev varsa, önce onu temizle.
	s.internalRemove(key)

	// Çalıştıktan sonra kendini temizleyecek özel bir wrapper oluştur.
	wrappedJob := func() {
		s.createJobWrapper(key)() // Normal wrapper'ı çalıştır
		s.internalRemove(key)     // İş bittikten sonra kendini sil
	}

	// Cron'un anlayacağı tek seferlik zaman formatını oluştur.
	spec := fmt.Sprintf("%d %d %d %d %d *", at.Second(), at.Minute(), at.Hour(), at.Day(), at.Month())

	entryID, err := s.cron.AddFunc(spec, wrappedJob)
	if err != nil {
		log.Printf("[AUTOMATION] ERROR scheduling one-off job for key '%s': %v", key, err)
		return err
	}

	// İşi kayıt defterine ekle.
	s.registry[key] = &jobInfo{
		Key:         key,
		IsRecurring: false,
		Job:         job,
		EntryID:     entryID,
		State:       StateIdle,
	}

	log.Printf("[AUTOMATION] Scheduled one-off job: '%s' for %s", key, at.Format(time.RFC3339))
	return nil
}

// Trigger, kayıtlı periyodik bir işi anahtarıyla manuel olarak ve güvenli bir şekilde tetikler.
func (s *AutomationService) Trigger(key string) error {
	s.mu.RLock()
	job, exists := s.registry[key]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("job with key '%s' not found in registry", key)
	}

	if !job.IsRecurring {
		return fmt.Errorf("cannot manually trigger a one-off scheduled job: '%s'", key)
	}

	log.Printf("[AUTOMATION] Manual trigger for job: '%s'", key)
	// Güvenli wrapper'ı yeni bir goroutine içinde çalıştır.
	go s.createJobWrapper(key)()

	return nil
}

// CancelSchedule, zamanlanmış tek seferlik bir görevi iptal eder.
func (s *AutomationService) CancelSchedule(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.internalRemove(key)
}

// internalRemove, kilitleme olmadan iptal/silme işlemini yapar.
func (s *AutomationService) internalRemove(key string) {
	if job, exists := s.registry[key]; exists {
		s.cron.Remove(job.EntryID)
		delete(s.registry, key)
		log.Printf("[AUTOMATION] Canceled/Removed job: '%s'", key)
	}
}

// createJobWrapper, bir işin güvenli çalışmasını sağlayan bir sarmalayıcı fonksiyon oluşturur.
// Bu fonksiyon, çakışmaları (race condition) ve panik durumlarını yönetir.
func (s *AutomationService) createJobWrapper(key string) JobFunc {
	return func() {
		s.mu.Lock()
		job, exists := s.registry[key]
		if !exists {
			s.mu.Unlock()
			log.Printf("[AUTOMATION] CRITICAL: Tried to run a job that no longer exists in registry: '%s'", key)
			return
		}

		if job.State == StateRunning {
			log.Printf("[AUTOMATION] SKIPPED job: '%s' is already running.", key)
			s.mu.Unlock()
			return
		}

		job.State = StateRunning
		s.mu.Unlock()

		log.Printf("[AUTOMATION] STARTED job: '%s'", key)
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[AUTOMATION] FATAL: RECOVERED from panic in job '%s': %v", key, r)
			}

			s.mu.Lock()
			// İş bittiğinde durumu tekrar "idle" yap.
			if job, exists := s.registry[key]; exists {
				job.State = StateIdle
			}
			s.mu.Unlock()
			log.Printf("[AUTOMATION] FINISHED job: '%s'", key)
		}()

		// Kayıt defterindeki asıl iş fonksiyonunu çalıştır.
		job.Job()
	}
}
