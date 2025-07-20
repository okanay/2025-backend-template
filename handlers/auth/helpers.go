package AuthHandler

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
)

// assembleLoginResponse, verilen bir userID için tüm kullanıcı verilerini (temel, detaylar, izinler)
// eş zamanlı olarak toplayıp standart bir LoginResponse formatında döndürür.
func (h *Handler) assembleLoginResponse(ctx context.Context, userID uuid.UUID) (*types.LoginResponse, error) {
	// Hata ve sonuçları taşımak için channel'lar.
	type userResult struct {
		user *types.User
		err  error
	}
	type detailsResult struct {
		details *types.UserDetails
		err     error
	}
	type permissionsResult struct {
		permissions []types.Permission
		err         error
	}

	userChan := make(chan userResult, 1)
	detailsChan := make(chan detailsResult, 1)
	permissionsChan := make(chan permissionsResult, 1)

	var wg sync.WaitGroup
	wg.Add(3)

	// Goroutine 1: Ana kullanıcı verisini çek.
	go func() {
		defer wg.Done()
		user, err := h.UserRepository.SelectByID(ctx, userID)
		userChan <- userResult{user: user, err: err}
	}()

	// Goroutine 2: Kullanıcı detaylarını çek.
	go func() {
		defer wg.Done()
		details, err := h.UserRepository.SelectUserDetailsByID(ctx, userID)
		detailsChan <- detailsResult{details: details, err: err}
	}()

	// Goroutine 3: Kullanıcı izinlerini çek.
	go func() {
		defer wg.Done()
		permissions, err := h.UserRepository.SelectPermissionsByUserID(ctx, userID)
		permissionsChan <- permissionsResult{permissions: permissions, err: err}
	}()

	wg.Wait()

	userRes := <-userChan
	detailsRes := <-detailsChan
	permsRes := <-permissionsChan

	if userRes.err != nil {
		return nil, fmt.Errorf("kullanıcı bilgileri alınamadı: %w", userRes.err)
	}
	// Diğer hatalar kritik değil, response'u etkilemez. Sadece loglanabilir.
	if detailsRes.err != nil { /* log */
	}
	if permsRes.err != nil { /* log */
	}

	if userRes.user.Status != types.UserStatusActive {
		return nil, fmt.Errorf("kullanıcı hesabı aktif değil")
	}

	// Sonucu birleştir.
	userView := types.UserView{
		ID:            userRes.user.ID,
		Role:          userRes.user.Role,
		Email:         userRes.user.Email,
		EmailVerified: userRes.user.EmailVerified,
	}

	if detailsRes.details != nil {
		userView.DisplayName = detailsRes.details.DisplayName
		userView.AvatarURL = detailsRes.details.AvatarURL
	}

	response := &types.LoginResponse{
		User:        userView,
		Permissions: permsRes.permissions,
	}

	return response, nil
}
