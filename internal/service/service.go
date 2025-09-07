package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gocha/internal/config"
	"gocha/internal/entity"
	"gocha/internal/repo"
	"gocha/pkg/gocha"

	"github.com/rs/zerolog"
)

var ErrPetNotFound = errors.New("питомец не найден")

type Service struct {
	cfg    *config.Configuration
	logger *zerolog.Logger
	repo   repo.Repository

	// Управление мониторингом
	monitoringChats map[int]context.CancelFunc // chatID -> cancel function
	monitoringMutex sync.RWMutex
}

func NewService(cfg *config.Configuration, logger *zerolog.Logger, repo repo.Repository) *Service {
	return &Service{
		cfg:             cfg,
		logger:          logger,
		repo:            repo,
		monitoringChats: make(map[int]context.CancelFunc),
	}
}

func (s *Service) NewPet(ctx context.Context, chatID int, name string) (*entity.Pet, error) {
	s.logger.Trace().Msg("create pet")
	extPet := gocha.NewPet(name)
	pet := GochaToPetEntity(extPet)

	err := s.repo.NewPet(ctx, pet, chatID)
	if err != nil {
		return nil, err
	}

	// Запускаем мониторинг для нового питомца
	s.startMonitoringForChat(ctx, chatID)

	return pet, nil
}

func (s *Service) PetFeed(ctx context.Context, chatID int) (entity.PetActionResult, error) {
	s.logger.Trace().Msg("pet feed")

	return s.petAction(ctx, chatID, func(p *gocha.Pet) gocha.Result {
		return p.Feed()
	})
}

func (s *Service) PetHeal(ctx context.Context, chatID int) (entity.PetActionResult, error) {
	s.logger.Trace().Msg("pet heal")

	return s.petAction(ctx, chatID, func(p *gocha.Pet) gocha.Result {
		return p.Heal()
	})
}

func (s *Service) PetClean(ctx context.Context, chatID int) (entity.PetActionResult, error) {
	s.logger.Trace().Msg("pet clean")

	return s.petAction(ctx, chatID, func(p *gocha.Pet) gocha.Result {
		return p.Clean()
	})
}

func (s *Service) PetPlay(ctx context.Context, chatID int) (entity.PetActionResult, error) {
	s.logger.Trace().Msg("pet play")

	return s.petAction(ctx, chatID, func(p *gocha.Pet) gocha.Result {
		return p.Play()
	})
}

func (s *Service) PetSleep(ctx context.Context, chatID int) (entity.PetActionResult, error) {
	s.logger.Trace().Msg("pet sleep")

	return s.petAction(ctx, chatID, func(p *gocha.Pet) gocha.Result {
		return p.Sleep()
	})
}

func (s *Service) PetWakeUp(ctx context.Context, chatID int) (entity.PetActionResult, error) {
	s.logger.Trace().Msg("pet wake up")

	return s.petAction(ctx, chatID, func(p *gocha.Pet) gocha.Result {
		return p.WakeUp()
	})
}

func (s *Service) PetBuru(ctx context.Context, chatID int) (entity.PetActionResult, error) {
	s.logger.Trace().Msg("pet buru")

	return s.petAction(ctx, chatID, func(p *gocha.Pet) gocha.Result {
		return p.WakeUp()
	})
}

func (s *Service) petAction(ctx context.Context, chatID int, action func(*gocha.Pet) gocha.Result) (entity.PetActionResult, error) {
	pet, err := s.repo.LoadPet(ctx, chatID)
	if err != nil {
		return entity.PetActionResult{
			Pet: nil,
			Result: entity.Result{
				Success: false,
				Message: "Не удалось загрузить питомца: " + err.Error(),
			},
		}, err
	}

	extPet := PetEntityToGocha(pet)

	result := action(extPet)

	pet = GochaToPetEntity(extPet)

	pet.GetAvatar(s.cfg.BaseUrl)

	err = s.SavePet(ctx, pet, chatID)
	if err != nil {
		return entity.PetActionResult{
			Pet: nil,
			Result: entity.Result{
				Success: false,
				Message: "Не удалось cохранить питомца: " + err.Error(),
			},
		}, err
	}

	return entity.PetActionResult{
		Pet: pet,
		Result: entity.Result{
			Success: result.Success,
			Message: result.Message,
		},
		Avatar: pet.Avatar,
	}, nil
}

func (s *Service) LoadPet(ctx context.Context, chatID int) (*entity.Pet, error) {
	s.logger.Trace().Msg("load pet")
	pet, err := s.repo.LoadPet(ctx, chatID)
	if err != nil {
		if errors.Is(err, repo.ErrPetNotFound) {
			s.logger.Warn().Msgf("Питомец не найден для chat_id: %d", chatID)

			return nil, ErrPetNotFound
		}

		return nil, err
	}

	return pet, nil
}

func (s *Service) SavePet(ctx context.Context, p *entity.Pet, chatID int) error {
	s.logger.Trace().Msg("save pet")
	err := s.repo.SavePet(ctx, p, chatID)
	if err != nil {
		return err
	}

	return nil
}

// startMonitoringForChat запускает мониторинг для конкретного чата
func (s *Service) startMonitoringForChat(ctx context.Context, chatID int) {
	s.monitoringMutex.Lock()
	defer s.monitoringMutex.Unlock()

	// Если мониторинг уже активен, останавливаем старый
	if cancelFunc, exists := s.monitoringChats[chatID]; exists {
		cancelFunc()
		s.logger.Debug().Msgf("Stopped existing monitoring for chat_id: %d", chatID)
	}

	// Создаем новый контекст для мониторинга этого чата
	monitorCtx, cancel := context.WithCancel(ctx)
	s.monitoringChats[chatID] = cancel

	// Запускаем мониторинг
	go s.MonitorAndLivePetAny(monitorCtx, chatID)
	s.logger.Info().Msgf("Started monitoring for chat_id: %d", chatID)
}

// stopMonitoringForChat останавливает мониторинг для конкретного чата
func (s *Service) stopMonitoringForChat(chatID int) {
	s.monitoringMutex.Lock()
	defer s.monitoringMutex.Unlock()

	if cancelFunc, exists := s.monitoringChats[chatID]; exists {
		cancelFunc()
		delete(s.monitoringChats, chatID)
		s.logger.Info().Msgf("Stopped monitoring for chat_id: %d", chatID)
	}
}

func (s *Service) MonitorPetsAll(ctx context.Context) error {
	s.logger.Trace().Msg("monitor pets all")

	chats, err := s.repo.GetChats(ctx)
	if err != nil {
		return fmt.Errorf("can't get chats: %w", err)
	}

	for _, chatID := range chats {
		s.startMonitoringForChat(ctx, chatID)
	}

	return nil
}

func (s *Service) MonitorAndLivePetAny(ctx context.Context, chatID int) {
	ticker := time.NewTicker(time.Duration(s.cfg.UpdateInterval) * time.Minute)
	defer ticker.Stop()

	s.logger.Trace().Msgf("Monitoring pet for chat_id: %d", chatID)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info().Msgf("Monitoring stopped for chat_id: %d", chatID)
			return
		case <-ticker.C:
			pet, err := s.repo.LoadPet(ctx, chatID)
			if err != nil {
				if errors.Is(err, repo.ErrPetNotFound) {
					s.logger.Warn().Msgf("Pet not found for chat_id: %d, stopping monitoring", chatID)
					s.stopMonitoringForChat(chatID)
					return
				}
				s.logger.Error().Err(err).Msg("can't load pet")
				continue
			}

			// Обновляем состояние питомца
			extPet := PetEntityToGocha(pet)
			extPet.DegradeOverTime(pet.LastUpdated)
			pet = GochaToPetEntity(extPet)

			// Сохраняем обновленное состояние
			if err := s.SavePet(ctx, pet, chatID); err != nil {
				s.logger.Error().Err(err).Msg("can't save pet")
				continue
			}

			// Проверяем и отправляем предупреждения
			now := time.Now()
			s.sendWarningIfNeeded(chatID, "health", pet.Health <= 20, "⚠️ Внимание! Здоровье питомца на критическом уровне!", now)
			s.sendWarningIfNeeded(chatID, "hunger", pet.Hunger >= 80, "⚠️ Внимание! Питомец очень голоден!", now)
			s.sendWarningIfNeeded(chatID, "happiness", pet.Happiness <= 20, "⚠️ Внимание! Питомец очень несчастен!", now)
			s.sendWarningIfNeeded(chatID, "energy", pet.Energy <= 20, "⚠️ Внимание! У питомца очень мало энергии!", now)
			s.sendWarningIfNeeded(chatID, "hygiene", pet.Hygiene <= 20, "⚠️ Внимание! Питомец очень грязный!", now)
		}
	}
}

func (s *Service) sendWarningIfNeeded(chatID int, alertType string, condition bool, message string, now time.Time) {
	if !condition {
		return
	}

	// Запрашиваем время последнего предупреждения из БД
	// lastAlert, err := s.repo.GetLastAlert(chatID, alertType)
	// if err != nil {
	//	s.logger.Error().Err(err).Msg("Ошибка получения последнего предупреждения")
	//	return
	//}

	// Проверяем, прошло ли достаточно времени с момента последнего предупреждения
	// if now.Sub(lastAlert) > time.Duration(s.cfg.AlertCooldown)*time.Minute {
	//	_, _ = s.bot.SendMessage(tu.Message(tu.ID(int64(chatID)), message))
	//	_ = s.repo.UpdateLastAlert(chatID, alertType, now) // Обновляем в БД
	//}
}

// Graceful shutdown - останавливает все мониторинги
func (s *Service) Stop() {
	s.monitoringMutex.Lock()
	defer s.monitoringMutex.Unlock()

	for chatID, cancelFunc := range s.monitoringChats {
		cancelFunc()
		s.logger.Info().Msgf("Stopped monitoring for chat_id: %d during shutdown", chatID)
	}
	s.monitoringChats = make(map[int]context.CancelFunc)
}

func PetEntityToGocha(pet *entity.Pet) *gocha.Pet {
	outPet := &gocha.Pet{
		Name:           pet.Name,
		Health:         pet.Health,
		Hunger:         pet.Hunger,
		Happiness:      pet.Happiness,
		Energy:         pet.Energy,
		Hygiene:        pet.Hygiene,
		State:          gocha.State(pet.State),
		SleepStartTime: pet.SleepStartTime,
	}
	outPet.EditConfig(gocha.Config{
		HungerDecayRate:    pet.Config.HungerDecayRate,
		EnergyDecayRate:    pet.Config.EnergyDecayRate,
		HygieneDecayRate:   pet.Config.HygieneDecayRate,
		HappinessDecayRate: pet.Config.HappinessDecayRate,
	})

	return outPet
}

func GochaToPetEntity(pet *gocha.Pet) *entity.Pet {
	cfg := pet.GetConfig()

	return &entity.Pet{
		Name:           pet.Name,
		Health:         pet.Health,
		Hunger:         pet.Hunger,
		Happiness:      pet.Happiness,
		Energy:         pet.Energy,
		Hygiene:        pet.Hygiene,
		State:          entity.State(pet.State),
		SleepStartTime: pet.SleepStartTime,
		Config: entity.PetConfig{
			HungerDecayRate:    cfg.HungerDecayRate,
			EnergyDecayRate:    cfg.EnergyDecayRate,
			HygieneDecayRate:   cfg.HygieneDecayRate,
			HappinessDecayRate: cfg.HappinessDecayRate,
		},
		LastUpdated: time.Now(),
	}
}
