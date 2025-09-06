package gocha

import (
	"fmt"
	"time"
)

const (
	defaultHungerDecayRate    = 2
	defaultEnergyDecayRate    = 3
	defaultHygieneDecayRate   = 1
	defaultHappinessDecayRate = 1
	defaultCoefficient        = 5
	MinStatValue              = 0
	MaxStatValue              = 100
)

type State string

const (
	Alive    State = "alive"
	Dead     State = "dead"
	Sleeping State = "sleeping"
)

const (
	PetIsDeadMessage = "Ваш питомец мертв."
)

type Result struct {
	Success bool
	Message string
}

type Pet struct {
	Name           string
	Health         int // Здоровье питомца в процентах.
	Hunger         int // Голод питомца в процентах.
	Happiness      int // Счастье питомца в процентах.
	Energy         int // Энергия питомца в процентах.
	Hygiene        int // Гигиена питомца в процентах.
	State          State
	SleepStartTime time.Time
	config         Config
}

type Config struct {
	HungerDecayRate    int
	EnergyDecayRate    int
	HygieneDecayRate   int
	HappinessDecayRate int
}

func NewPet(name string) *Pet {
	config := Config{
		HungerDecayRate:    defaultHungerDecayRate,
		EnergyDecayRate:    defaultEnergyDecayRate,
		HygieneDecayRate:   defaultHygieneDecayRate,
		HappinessDecayRate: defaultHappinessDecayRate,
	}

	return &Pet{
		Name:      name,
		Health:    MaxStatValue,
		Hunger:    MaxStatValue,
		Happiness: MaxStatValue,
		Energy:    MaxStatValue,
		Hygiene:   MaxStatValue,
		State:     Alive,
		config:    config,
	}
}

func (p *Pet) EditConfig(cfg Config) {
	p.config = cfg
}

func (p *Pet) GetConfig() Config {
	return p.config
}

func (p *Pet) Feed() Result {
	if p.IsDead() {
		p.Kill()

		return Result{Success: false, Message: PetIsDeadMessage}
	}

	p.Hunger += defaultCoefficient // Уменьшаем голод

	if p.IsOverfed() {
		p.Health = clamp(p.Health-defaultCoefficient, MinStatValue, MaxStatValue)
		if p.Health == MinStatValue {
			p.Kill()

			return Result{Success: false, Message: "Питомец умер из-за перекорма!"}
		}

		p.Hunger = clamp(p.Hunger, MinStatValue, MaxStatValue)

		return Result{Success: false, Message: "Питомец перекормлен! Здоровье ухудшилось."}
	}

	p.Hunger = clamp(p.Hunger, MinStatValue, MaxStatValue)

	return Result{Success: true, Message: "Питомец покормлен!"}
}

func (p *Pet) Heal() Result {
	if p.IsDead() {
		p.Kill()

		return Result{Success: false, Message: PetIsDeadMessage}
	}

	if p.IsOverHealed() {
		if p.Energy == MinStatValue {
			return Result{Success: false, Message: "Питомец слишком устал, чтобы лечиться!"}
		}

		p.Energy = clamp(p.Energy-defaultCoefficient, MinStatValue, MaxStatValue)
		p.Happiness = clamp(p.Happiness-defaultCoefficient, MinStatValue, MaxStatValue)

		return Result{Success: true, Message: fmt.Sprintf("Питомец перелечен! Энергия: -%d", defaultCoefficient)}
	}

	p.Health += defaultCoefficient
	p.Health = clamp(p.Health, MinStatValue, MaxStatValue)

	if p.Health == MaxStatValue {
		return Result{Success: true, Message: "Питомец полностью здоров!"}
	}

	return Result{Success: true, Message: fmt.Sprintf("Питомца полечили. Здоровье: +%d", defaultCoefficient)}
}

func (p *Pet) Play() Result {
	if p.IsDead() {
		p.Kill()

		return Result{Success: false, Message: PetIsDeadMessage}
	}

	if p.Energy < defaultCoefficient {
		p.Happiness = clamp(p.Happiness-defaultCoefficient, MinStatValue, MaxStatValue)
		p.Health = clamp(p.Health-defaultCoefficient/2, MinStatValue, MaxStatValue)

		if p.Health == MinStatValue {
			p.Kill()

			return Result{Success: false, Message: "Питомец умер."}
		}
	} else {
		p.Happiness = clamp(p.Happiness+defaultCoefficient, MinStatValue, MaxStatValue)
		p.Energy = clamp(p.Energy-defaultCoefficient/2, MinStatValue, MaxStatValue)
	}

	return Result{
		Success: true,
		Message: fmt.Sprintf("Питомец играл. Счастье: +%d, Энергия: -%d", defaultCoefficient, defaultCoefficient/2),
	}
}

func (p *Pet) Clean() Result {
	if p.IsDead() {
		p.Kill()

		return Result{Success: false, Message: PetIsDeadMessage}
	}

	p.Hygiene = clamp(p.Hygiene+defaultCoefficient, MinStatValue, MaxStatValue)

	if p.Hygiene == MaxStatValue {
		return Result{Success: true, Message: "Питомец полностью чист!"}
	}

	return Result{
		Success: true,
		Message: fmt.Sprintf("Питомца помыли. Гигиена: +%d", defaultCoefficient),
	}
}

func (p *Pet) Sleep() Result {
	if p.IsDead() {
		p.Kill()

		return Result{Success: false, Message: PetIsDeadMessage}
	}

	if p.IsSleeping() {
		return Result{
			Success: false,
			Message: "Питомец уже спит.",
		}
	}

	p.State = Sleeping
	p.SleepStartTime = time.Now()

	return Result{
		Success: true,
		Message: "Питомца отправили спать.",
	}
}

func (p *Pet) WakeUp() Result {
	if p.IsDead() {
		p.Kill()

		return Result{Success: false, Message: PetIsDeadMessage}
	}

	if !p.IsSleeping() {
		return Result{Success: false, Message: "Питомец не спит."}
	}

	// Вычисляем продолжительность сна
	sleepDuration := time.Since(p.SleepStartTime)
	minutesSlept := int(sleepDuration.Minutes())

	// Минимальное время сна
	if minutesSlept < 1 {
		return Result{Success: false, Message: "Питомец не выспался."}
	}

	// Максимальное время сна
	if minutesSlept > 480 { // 8 часов
		minutesSlept = 480
	}

	// Вычисляем изменения
	energyGained := minutesSlept * 2 // 2 единицы энергии за каждую минуту сна
	hungerGained := minutesSlept * 1 // 1 единица голода за каждую минуту сна

	// Применяем изменения, но проверяем их границы
	newEnergy := clamp(p.Energy+energyGained, MinStatValue, MaxStatValue)
	newHunger := clamp(p.Hunger+hungerGained, MinStatValue, MaxStatValue)

	// Если ничего не изменилось, результат бессмысленный
	if newEnergy == p.Energy && newHunger == p.Hunger {
		p.State = Alive

		return Result{Success: false, Message: "Питомец не получил пользы от сна."}
	}

	// Обновляем состояние питомца
	p.Energy = newEnergy
	p.Hunger = newHunger

	// Меняем состояние на "бодрствует"
	p.State = Alive

	// Формируем сообщение с результатами
	message := fmt.Sprintf(
		"Питомец проснулся! Спал %d минут. Энергия +%d, голод +%d.",
		minutesSlept, energyGained, hungerGained,
	)

	return Result{Success: true, Message: message}
}

func (p *Pet) DegradeOverTime(lastUpdated time.Time) {
	minutes := int(time.Since(lastUpdated).Minutes())

	if minutes <= 0 {
		return
	}

	if p.State == Sleeping {
		p.updateSleepingState(minutes)
	} else {
		p.updateAwakeState(minutes)
	}

	p.applyDamage(minutes)

	if p.Health == MinStatValue {
		p.Kill()
	}
}

func (p *Pet) Kill() {
	p.State = Dead
	p.Health = MinStatValue
	p.Hunger = MinStatValue
	p.Happiness = MinStatValue
	p.Energy = MinStatValue
	p.Hygiene = MinStatValue
}

func (p *Pet) updateSleepingState(minutes int) {
	p.Energy = clamp(p.Energy+minutes*defaultCoefficient, MinStatValue, MaxStatValue)
	p.Hunger = clamp(p.Hunger-minutes*2, MinStatValue, MaxStatValue)
}

func (p *Pet) updateAwakeState(minutes int) {
	p.Hunger = clamp(p.Hunger-minutes*p.config.HungerDecayRate, MinStatValue, MaxStatValue)
	p.Energy = clamp(p.Energy-minutes*p.config.EnergyDecayRate, MinStatValue, MaxStatValue)
	p.Hygiene = clamp(p.Hygiene-minutes*p.config.HygieneDecayRate, MinStatValue, MaxStatValue)
	p.Happiness = clamp(p.Happiness-minutes*p.config.HappinessDecayRate, MinStatValue, MaxStatValue)
}

func (p *Pet) applyDamage(minutes int) {
	damage := 0

	if p.Hunger == MinStatValue || p.Hygiene == MinStatValue || p.Energy == MinStatValue || p.IsUnhappy() {
		damage += minutes
	}

	p.Health = max(p.Health-damage, MinStatValue)
}

func clamp(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}

	if value > maxValue {
		return maxValue
	}

	return value
}
