package entity

import (
	"fmt"
	"time"
)

type State string

const (
	PetAlive    State = "alive"
	PetDead     State = "dead"
	PetSleeping State = "sleeping"
)

// UIConfig Конфигурация для UI.
type UIConfig struct {
	CriticalThreshold int `json:"criticalThreshold"` // 20
	WarningThreshold  int `json:"warningThreshold"`  // 40
	GoodThreshold     int `json:"goodThreshold"`     // 60
}

// PetStatus Расширенная информация о состоянии питомца.
type PetStatus struct {
	AverageStats     int    `json:"averageStats"`
	IsCritical       bool   `json:"isCritical"`
	IsWarning        bool   `json:"isWarning"`
	StatusMessage    string `json:"statusMessage"`
	StatusType       string `json:"statusType"` // "good", "warning", "danger"
	CanPerformAction bool   `json:"canPerformAction"`
}

// AvailableActions Информация о доступных действиях.
type AvailableActions struct {
	CanFeed   bool `json:"canFeed"`
	CanPlay   bool `json:"canPlay"`
	CanClean  bool `json:"canClean"`
	CanHeal   bool `json:"canHeal"`
	CanSleep  bool `json:"canSleep"`
	CanWakeUp bool `json:"canWakeUp"`
}

type Pet struct {
	Name             string           `json:"name"`
	Health           int              `json:"health"`
	Hunger           int              `json:"hunger"`
	Happiness        int              `json:"happiness"`
	Energy           int              `json:"energy"`
	Hygiene          int              `json:"hygiene"`
	State            State            `json:"state"`
	SleepStartTime   time.Time        `json:"sleepStartTime"`
	Config           PetConfig        `json:"config"`
	LastUpdated      time.Time        `json:"lastUpdated"`
	Age              int              `json:"age"`
	Avatar           Avatar           `json:"avatar"`
	Status           PetStatus        `json:"status"`
	AvailableActions AvailableActions `json:"availableActions"`
	UIConfig         UIConfig         `json:"uiConfig"`
}

type PetConfig struct {
	HungerDecayRate    int
	EnergyDecayRate    int
	HygieneDecayRate   int
	HappinessDecayRate int
}

// UpdateStatus Метод для вычисления расширенного статуса.
func (pet *Pet) UpdateStatus() {
	// Вычисляем средние статы
	avg := (pet.Hunger + pet.Happiness + pet.Hygiene + pet.Health + pet.Energy) / 5
	pet.Status.AverageStats = avg

	// Определяем критическое состояние
	pet.Status.IsCritical = avg <= 20 || pet.Health <= 20
	pet.Status.IsWarning = avg <= 40 || pet.Health <= 40

	// Определяем возможность выполнения действий
	pet.Status.CanPerformAction = pet.State != PetDead

	// Генерируем статусное сообщение
	pet.generateStatusMessage()

	// Определяем доступные действия
	pet.updateAvailableActions()

	// Устанавливаем UI конфигурацию
	pet.UIConfig = UIConfig{
		CriticalThreshold: 20,
		WarningThreshold:  40,
		GoodThreshold:     60,
	}
}

// Генерация статусного сообщения.
func (pet *Pet) generateStatusMessage() {
	switch pet.State {
	case PetDead:
		pet.Status.StatusMessage = "💀 Питомец умер... Создайте нового!"
		pet.Status.StatusType = "danger"
	case PetSleeping:
		pet.Status.StatusMessage = "💤 Питомец спит..."
		pet.Status.StatusType = "good"
	default:
		// Определяем статус по состоянию здоровья и статам
		if pet.Health <= 20 {
			pet.Status.StatusMessage = "🤒 Питомец болен!"
			pet.Status.StatusType = "danger"
		} else if pet.Energy <= 20 {
			pet.Status.StatusMessage = "😴 Питомец устал!"
			pet.Status.StatusType = "warning"
		} else if pet.Status.AverageStats <= 30 {
			pet.Status.StatusMessage = "😿 Питомцу грустно..."
			pet.Status.StatusType = "warning"
		} else if pet.Status.AverageStats >= 80 {
			pet.Status.StatusMessage = "😸 Питомец счастлив!"
			pet.Status.StatusType = "good"
		} else {
			pet.Status.StatusMessage = "😊 Всё в порядке"
			pet.Status.StatusType = "good"
		}
	}
}

// Определение доступных действий.
func (pet *Pet) updateAvailableActions() {
	isDead := pet.State == PetDead
	isSleeping := pet.State == PetSleeping

	pet.AvailableActions = AvailableActions{
		CanFeed:   !isDead && !isSleeping && pet.Hunger < 100,
		CanPlay:   !isDead && !isSleeping && pet.Energy > 20 && pet.Happiness < 100,
		CanClean:  !isDead && !isSleeping && pet.Hygiene < 100,
		CanHeal:   !isDead && !isSleeping && pet.Health < 100,
		CanSleep:  !isDead && !isSleeping && pet.Energy <= 30,
		CanWakeUp: !isDead && isSleeping,
	}
}

func (pet *Pet) GetAvatar(baseURL string) {
	// Определяем состояние
	if pet.State == PetDead {
		pet.Avatar = Avatar{
			Image: baseURL + "/dead.png",
			Emoji: "💀",
			Mood:  "💀",
		}

		return
	}

	if pet.State == PetSleeping {
		pet.Avatar = Avatar{
			Image: baseURL + "/sleeping.png",
			Emoji: "😴",
			Mood:  "💤",
		}

		return
	}

	// Считаем среднее по основным статам
	avg := (pet.Hunger + pet.Happiness + pet.Hygiene + pet.Health + pet.Energy) / 5

	var img, emoji, moodEmoji string

	if pet.Health <= 20 {
		img = baseURL + "/sick.png"
		emoji = "🤒"
		moodEmoji = "🤒"
	} else if pet.Energy <= 20 {
		img = baseURL + "/tired.png"
		emoji = "😴"
		moodEmoji = "😴"
	} else if avg >= 80 {
		img = baseURL + "/happy.png"
		emoji = "😸"
		moodEmoji = "😸"
	} else if avg >= 60 {
		img = baseURL + "/default.png"
		emoji = "🐱"
		moodEmoji = "😊"
	} else if avg >= 40 {
		img = baseURL + "/default.png"
		emoji = "🐱"
		moodEmoji = "😐"
	} else {
		img = baseURL + "/sad.png"
		emoji = "🙀"
		moodEmoji = "😿"
	}

	pet.Avatar = Avatar{
		Image: img,
		Emoji: emoji,
		Mood:  moodEmoji,
	}

	pet.UpdateStatus()
}

// CanPerformAction Метод для проверки возможности выполнения конкретного действия.
func (pet *Pet) CanPerformAction(action string) (bool, string) {
	if pet.State == PetDead {
		return false, "Питомец мертв"
	}

	switch action {
	case "feed":
		if pet.State == PetSleeping {
			return false, "Питомец спит"
		}
		if pet.Hunger >= 100 {
			return false, "Питомец не голоден"
		}

		return true, ""

	case "play":
		if pet.State == PetSleeping {
			return false, "Питомец спит"
		}
		if pet.Energy <= 20 {
			return false, "Питомец слишком устал"
		}
		if pet.Happiness >= 100 {
			return false, "Питомец уже счастлив"
		}
		return true, ""

	case "clean":
		if pet.State == PetSleeping {
			return false, "Питомец спит"
		}
		if pet.Hygiene >= 100 {
			return false, "Питомец уже чистый"
		}
		return true, ""

	case "heal":
		if pet.State == PetSleeping {
			return false, "Питомец спит"
		}
		if pet.Health >= 100 {
			return false, "Питомец здоров"
		}
		return true, ""

	case "sleep":
		if pet.State == PetSleeping {
			return false, "Питомец уже спит"
		}
		if pet.Energy > 30 {
			return false, "Питомец не устал"
		}
		return true, ""

	case "wakeup":
		if pet.State != PetSleeping {
			return false, "Питомец не спит"
		}
		return true, ""

	default:
		return false, "Неизвестное действие"
	}
}

// PetActionResult Обновленная структура результата действия.
type PetActionResult struct {
	Pet            *Pet   `json:"pet"`
	Result         Result `json:"result"`
	Avatar         Avatar `json:"avatar"`
	ActionFeedback string `json:"actionFeedback"` // Специальное сообщение для действия
}

func (r *PetActionResult) GetAvatar(baseURL string) {
	if r.Pet == nil {
		return
	}
	r.Pet.GetAvatar(baseURL)
	r.Avatar = r.Pet.Avatar
}

// GenerateActionFeedback Генерация обратной связи для действия.
func (r *PetActionResult) GenerateActionFeedback(action string) {
	actionNames := map[string]string{
		"feed":   "покормили",
		"play":   "поиграли",
		"clean":  "помыли",
		"heal":   "вылечили",
		"sleep":  "уложили спать",
		"wakeup": "разбудили",
	}

	name := actionNames[action]
	if name == "" {
		name = action
	}

	if r.Result.Success {
		r.ActionFeedback = fmt.Sprintf("🎉 Успешно %s!", name)
	} else {
		r.ActionFeedback = fmt.Sprintf("❌ Не удалось %s", name)
	}
}

// Messages Локализация сообщений.
type Messages struct {
	PetDead       string `json:"petDead"`
	PetSleeping   string `json:"petSleeping"`
	PetSick       string `json:"petSick"`
	PetTired      string `json:"petTired"`
	PetSad        string `json:"petSad"`
	PetHappy      string `json:"petHappy"`
	PetOk         string `json:"petOk"`
	NoInitData    string `json:"noInitData"`
	PetNotFound   string `json:"petNotFound"`
	CreateSuccess string `json:"createSuccess"`
}

func GetMessages(lang string) Messages {
	// Можно загружать из файлов конфигурации или базы данных
	switch lang {
	case "en":
		return Messages{
			PetDead:       "💀 Pet is dead... Create a new one!",
			PetSleeping:   "💤 Pet is sleeping...",
			PetSick:       "🤒 Pet is sick!",
			PetTired:      "😴 Pet is tired!",
			PetSad:        "😿 Pet is sad...",
			PetHappy:      "😸 Pet is happy!",
			PetOk:         "😊 Everything is fine",
			NoInitData:    "No initData",
			PetNotFound:   "Pet not found",
			CreateSuccess: "🎉 Congratulations! Pet %s created!",
		}
	default: // ru
		return Messages{
			PetDead:       "💀 Питомец умер... Создайте нового!",
			PetSleeping:   "💤 Питомец спит...",
			PetSick:       "🤒 Питомец болен!",
			PetTired:      "😴 Питомец устал!",
			PetSad:        "😿 Питомцу грустно...",
			PetHappy:      "😸 Питомец счастлив!",
			PetOk:         "😊 Всё в порядке",
			NoInitData:    "Нет initData",
			PetNotFound:   "Питомец не найден",
			CreateSuccess: "🎉 Поздравляем! Питомец %s создан!",
		}
	}
}

type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type Avatar struct {
	Image string `json:"image"`
	Emoji string `json:"emoji"` // Основной эмодзи аватара
	Mood  string `json:"mood"`
}
