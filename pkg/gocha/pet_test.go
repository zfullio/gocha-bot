package gocha

import (
	"testing"
	"time"
)

func TestPet_Feed(t *testing.T) {
	t.Parallel()

	t.Run("кормить сытого питомца", func(t *testing.T) {
		p := NewPet("")
		p.Feed()

		if p.Hunger != 0 {
			t.Errorf("Feed() = %v, want 0", p.Hunger)
		}
	})

	t.Run("кормить голодного питомца", func(t *testing.T) {
		p := NewPet("")
		p.Hunger = 100
		p.Feed()

		if p.Hunger != 90 {
			t.Errorf("Feed() = %v, want 90", p.Hunger)
		}
	})

	t.Run("кормить питомца с Hunger = 0 несколько раз", func(t *testing.T) {
		p := NewPet("")
		p.Feed()
		p.Feed()

		if p.Hunger != 0 {
			t.Errorf("Feed() = %v, want 0", p.Hunger)
		}
	})

	t.Run("кормить мертвого питомца", func(t *testing.T) {
		p := NewPet("")
		p.State = Dead
		p.Feed()

		if p.Hunger != 100 {
			t.Errorf("Feed() should not change Hunger when dead")
		}
	})
}

func TestPet_Heal(t *testing.T) {
	t.Parallel()

	t.Run("лечить здорового питомца", func(t *testing.T) {
		p := NewPet("")
		p.Heal()

		if p.Health != 100 {
			t.Errorf("Heal() Health = %v, want 100", p.Health)
		}
	})

	t.Run("лечить питомца с Health = 95", func(t *testing.T) {
		p := NewPet("")
		p.Health = 95
		p.Heal()

		if p.Health > 100 {
			t.Errorf("Heal() Health = %v, should not exceed 100", p.Health)
		}
	})

	t.Run("лечить мертвого питомца", func(t *testing.T) {
		p := NewPet("")
		p.State = Dead
		p.Heal()

		if p.Health != 0 {
			t.Errorf("Heal() should not work on Dead pet, but got Health = %v", p.Health)
		}
	})

	t.Run("лечить больного питомца с нулевой энергией", func(t *testing.T) {
		p := NewPet("")
		p.Health = 50
		p.Energy = 0
		p.Heal()

		if p.Health != 60 {
			t.Errorf("Heal() Не увеличивается Health при нулевой энергии")
		}
	})
}

func TestPet_Play(t *testing.T) {
	t.Parallel()

	t.Run("играть с питомцем без энергии", func(t *testing.T) {
		p := NewPet("")
		p.Energy = 0
		p.Play()

		if p.Happiness != 90 {
			t.Errorf("Play() should not increase Happiness when Energy is 0")
		}
	})
}

func TestPet_Clean(t *testing.T) {
	t.Parallel()

	t.Run("чистить мертвого питомца", func(t *testing.T) {
		p := NewPet("")
		p.State = Dead
		p.Clean()

		if p.Hygiene != 0 {
			t.Errorf("Dead pet should not be cleaned")
		}
	})
}

func TestPet_Sleep(t *testing.T) {
	t.Parallel()

	t.Run("засыпание мертвого питомца", func(t *testing.T) {
		p := NewPet("")
		p.State = Dead
		p.Sleep()

		if p.State == Sleeping {
			t.Errorf("Dead pet should not be able to sleep")
		}
	})

	t.Run("просыпание питомца после 24 часов сна", func(t *testing.T) {
		p := NewPet("")
		p.State = Sleeping
		p.SleepStartTime = time.Now().Add(-24 * time.Hour)
		p.WakeUp()

		if p.Energy > 100 {
			t.Errorf("Energy should not exceed 100 after long sleep")
		}
	})
}

func TestPet_DegradeOverTime(t *testing.T) {
	t.Parallel()

	t.Run("деградация за 24 часа", func(t *testing.T) {
		p := NewPet("")
		lastUpdated := time.Now().Add(-24 * time.Hour)
		p.DegradeOverTime(lastUpdated)

		if p.Hunger > 100 {
			t.Errorf("Hunger exceeded 100: %v", p.Hunger)
		}

		if p.Energy < 0 {
			t.Errorf("Energy dropped below 0: %v", p.Energy)
		}
	})
}
