package gocha

func (p *Pet) IsDead() bool {
	return p.State == Dead
}

func (p *Pet) IsOverfed() bool {
	return p.Hunger < 0
}

func (p *Pet) IsOverHealed() bool {
	return p.Health == 100
}

func (p *Pet) IsDirty() bool {
	return p.Hygiene <= 20
}

func (p *Pet) IsUnhappy() bool {
	return p.Happiness <= 20
}

func (p *Pet) IsAlive() bool {
	return p.State == Alive
}

func (p *Pet) IsSleeping() bool {
	return p.State == Sleeping
}
