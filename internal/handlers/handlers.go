package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"gocha/internal/entity"
	"gocha/internal/service"

	"github.com/rs/zerolog"
	initdata "github.com/telegram-mini-apps/init-data-golang"
)

const (
	PetNotFindErr = "Питомец не найден"
)

type PetHandlers struct {
	s       *service.Service
	logger  zerolog.Logger
	baseUrl string
	isDev   bool
}

func NewPetHandlers(logger zerolog.Logger, s *service.Service, baseUrl string, isDev bool) *PetHandlers {
	return &PetHandlers{logger: logger, s: s, baseUrl: baseUrl, isDev: isDev}
}

func (h *PetHandlers) PetNewHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	h.logger.Info().Msg("pet create")

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Name string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Failed to decode request")

		return
	}

	tgData := r.Header.Get("X-Telegram-Init-Data")
	if tgData == "" {
		json.NewEncoder(w).Encode(entity.APIResponse[entity.Pet]{
			Success: false,
			Message: "Нет initData",
		})

		return
	}

	parseData, err := initdata.Parse(tgData)
	if err != nil {
		json.NewEncoder(w).Encode(entity.APIResponse[entity.Pet]{
			Success: false,
			Message: "Не удалось прочитать tg-init-data",
		})

		return
	}

	err = h.s.NewPet(ctx, getPetID(parseData), req.Name)
	if err != nil {
		json.NewEncoder(w).Encode(entity.APIResponse[entity.Pet]{
			Success: false,
			Message: "Не могу создать питомца",
		})

		return
	}

	json.NewEncoder(w).Encode(entity.APIResponse[entity.Pet]{
		Success: true,
		Message: fmt.Sprintf("🎉 Поздравляем! Питомец %s создан!", req.Name),
		Data:    entity.Pet{},
	})

}

func (h *PetHandlers) PetInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	w.Header().Set("Content-Type", "application/json")

	tgData := r.Header.Get("X-Telegram-Init-Data")
	if tgData == "" {
		json.NewEncoder(w).Encode(entity.APIResponse[entity.Pet]{
			Success: false,
			Message: "Нет initData",
		})

		return
	}

	parseData, err := initdata.Parse(tgData)
	if err != nil {
		json.NewEncoder(w).Encode(entity.APIResponse[entity.Pet]{
			Success: false,
			Message: "Не удалось прочитать tg-init-data",
		})

		return
	}

	pet, err := h.s.LoadPet(ctx, getPetID(parseData))
	if err != nil {
		if errors.Is(err, service.ErrPetNotFound) {
			json.NewEncoder(w).Encode(entity.APIResponse[entity.Pet]{
				Success: false,
				Message: PetNotFindErr,
			})
		} else {
			json.NewEncoder(w).Encode(entity.APIResponse[entity.Pet]{
				Success: false,
				Message: "Ошибка загрузки питомца",
			})
		}

		return
	}

	pet.GetAvatar(fmt.Sprintf("%s/%s", h.baseUrl, "static"))
	pet.UpdateStatus()

	json.NewEncoder(w).Encode(entity.APIResponse[entity.Pet]{
		Success: true,
		Data:    *pet, // возвращаем полную структуру питомца
	})
}

func (h *PetHandlers) handlePetAction(w http.ResponseWriter, r *http.Request, action func(ctx context.Context, petID int) (entity.PetActionResult, error), actionName string) {
	ctx := context.Background()
	w.Header().Set("Content-Type", "application/json")

	tgData := r.Header.Get("X-Telegram-Init-Data")
	if tgData == "" {
		json.NewEncoder(w).Encode(entity.APIResponse[entity.PetActionResult]{
			Success: false,
			Message: "Нет initData",
		})
		return
	}

	parseData, err := initdata.Parse(tgData)
	if err != nil {
		json.NewEncoder(w).Encode(entity.APIResponse[entity.PetActionResult]{
			Success: false,
			Message: "Не удалось прочитать tg-init-data",
		})
		return
	}

	petID := getPetID(parseData)

	// Сначала загружаем питомца для проверки возможности действия
	pet, err := h.s.LoadPet(ctx, petID)
	if err != nil {
		if errors.Is(err, service.ErrPetNotFound) {
			json.NewEncoder(w).Encode(entity.APIResponse[entity.PetActionResult]{
				Success: false,
				Message: PetNotFindErr,
			})
		} else {
			json.NewEncoder(w).Encode(entity.APIResponse[entity.PetActionResult]{
				Success: false,
				Message: "Ошибка загрузки питомца",
			})
		}
		return
	}

	pet.UpdateStatus()

	// Проверяем возможность выполнения действия
	canPerform, reason := pet.CanPerformAction(actionName)
	if !canPerform {
		json.NewEncoder(w).Encode(entity.APIResponse[entity.PetActionResult]{
			Success: false,
			Message: reason,
		})

		return
	}

	// Выполняем действие
	result, err := action(ctx, petID)
	if err != nil {
		if errors.Is(err, service.ErrPetNotFound) {
			json.NewEncoder(w).Encode(entity.APIResponse[entity.PetActionResult]{
				Success: false,
				Message: PetNotFindErr,
			})
		} else {
			json.NewEncoder(w).Encode(entity.APIResponse[entity.PetActionResult]{
				Success: false,
				Message: "Ошибка при выполнении действия",
			})
		}
		return
	}

	result.GetAvatar(fmt.Sprintf("%s/%s", h.baseUrl, "static"))
	result.GenerateActionFeedback(actionName)

	json.NewEncoder(w).Encode(entity.APIResponse[entity.PetActionResult]{
		Success: true,
		Data:    result,
	})
}

func (h *PetHandlers) PetFeedHandler(w http.ResponseWriter, r *http.Request) {
	h.handlePetAction(w, r, h.s.PetFeed, "feed")
}

func (h *PetHandlers) PetHealHandler(w http.ResponseWriter, r *http.Request) {
	h.handlePetAction(w, r, h.s.PetHeal, "heal")
}

func (h *PetHandlers) PetPlayHandler(w http.ResponseWriter, r *http.Request) {
	h.handlePetAction(w, r, h.s.PetPlay, "play")
}

func (h *PetHandlers) PetCleanHandler(w http.ResponseWriter, r *http.Request) {
	h.handlePetAction(w, r, h.s.PetClean, "clean")
}

func (h *PetHandlers) PetSleepHandler(w http.ResponseWriter, r *http.Request) {
	h.handlePetAction(w, r, h.s.PetSleep, "sleep")
}

func (h *PetHandlers) PetWakeUpHandler(w http.ResponseWriter, r *http.Request) {
	h.handlePetAction(w, r, h.s.PetWakeUp, "wakeup")
}

func (h *PetHandlers) DebugMockInitDataHandler(w http.ResponseWriter, r *http.Request) {
	if !h.isDev {
		http.Error(w, "Not available in production", http.StatusForbidden)

		return
	}

	h.logger.Info().Msg("serving mock init-data for dev")

	// Пример тестового initData (подпись не нужна — бэк сам проверит)
	userData := `{"id":123456789,"first_name":"TestUser","username":"testuser","language_code":"ru","is_bot":false}`
	values := url.Values{}
	values.Set("user", userData)
	values.Set("auth_date", "1735689600")
	values.Set("hash", "validhashfordevonly") // будет проверяться, но в dev можно игнорировать или подделать

	// Возвращаем как есть — строку
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"initData": values.Encode(),
	})
}

func (h *PetHandlers) DebugInitTgConfigHandler(w http.ResponseWriter, r *http.Request) {
	if !h.isDev {
		http.Error(w, "Not available in production", http.StatusForbidden)

		return
	}

	h.logger.Info().Msg("serving init-config for dev")

	// === initData (можно переиспользовать из DebugMockInitDataHandler) ===
	userData := `{"id":123456789,"first_name":"TestUser","username":"testuser","language_code":"ru","is_bot":false}`
	values := url.Values{}
	values.Set("user", userData)
	values.Set("auth_date", "1735689600")
	values.Set("hash", "validhashfordevonly")

	initData := values.Encode()

	// === Конфиг с методами в виде строк ===
	config := map[string]any{
		"initData":       initData,
		"platform":       "web",
		"version":        "6.0",
		"viewportHeight": 800,
		"themeParams": map[string]string{
			"bg_color":           "#ffffff",
			"text_color":         "#000000",
			"button_color":       "#6366f1",
			"secondary_bg_color": "#f0f0f0",
			"hint_color":         "#999999",
		},
		"methods": map[string]any{
			"ready":   "function() { console.log('Ready (mock)'); }",
			"expand":  "function() { console.log('Expanded (mock)'); }",
			"onEvent": "function(event, handler) { console.log('Event listener registered: ' + event); if (event === 'themeChanged') setTimeout(handler, 100); if (event === 'viewportChanged') setTimeout(handler, 100); }",
		},
		"mainButton": map[string]string{
			"setText":      "function(text) { console.log('MainButton text:', text); }",
			"show":         "function() { console.log('MainButton show'); }",
			"hide":         "function() { console.log('MainButton hide'); }",
			"onClick":      "function(fn) { console.log('MainButton onClick registered'); window._mainButtonClick = fn; }",
			"offClick":     "function() { window._mainButtonClick = null; }",
			"showProgress": "function() { console.log('MainButton showProgress'); }",
			"hideProgress": "function() { console.log('MainButton hideProgress'); }",
		},
		"hapticFeedback": map[string]string{
			"impactOccurred":       "function(style) { console.log('Haptic:', style); }",
			"notificationOccurred": "function(type) { console.log('Notify haptic:', type); }",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(config)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to encode init-config")
		http.Error(w, "failed to encode config", http.StatusInternalServerError)
	}
}

func (h *PetHandlers) respondWithError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)

	errorResponse := map[string]string{
		"success": "false",
		"message": message,
	}

	err := json.NewEncoder(w).Encode(errorResponse)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to send error response")
		// Фолбэк — уже статус установлен
	}
}

func getPetID(data initdata.InitData) int {
	if data.ChatInstance != 0 {
		return int(data.ChatInstance)
	}

	// 2. Если есть объект chat — значит, запуск из attachment menu (в группе/супергруппе/канале)
	if data.Chat.ID != 0 {
		return int(data.Chat.ID)
	}

	if data.User.ID != 0 {
		return int(data.User.ID)
	}

	// На всякий случай — если ничего не подошло, хотя такого быть не должно при валидных данных
	return 0
}
