package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"time"

	"gocha/internal/entity"
	"gocha/internal/repo"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

//go:embed sql/init.sql
var sqlInitDB string

//go:embed sql/new_pet.sql
var sqlNewPet string

//go:embed sql/save_pet.sql
var sqlSavePet string

//go:embed sql/load_pet.sql
var sqlLoadPet string

//go:embed sql/get_chats.sql
var sqlGetChats string

//go:embed sql/deactivate_pets.sql
var sqlDeactivatePets string

type Repository struct {
	logger *zerolog.Logger
	db     *pgxpool.Pool
}

func NewRepository(logger *zerolog.Logger, db *pgxpool.Pool) *Repository {
	ctx := context.Background()
	innerLoger := logger.With().Str("type", "postgres").Logger()

	_, err := db.Exec(ctx, sqlInitDB)
	if err != nil {
		logger.Fatal().Err(err).Msg("can't init database")
	}

	return &Repository{
		db:     db,
		logger: &innerLoger,
	}
}

func (r *Repository) NewPet(ctx context.Context, p *entity.Pet, chatID int) error {
	_, err := r.db.Exec(ctx, sqlDeactivatePets, chatID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sqlNewPet, chatID, p.Name, p.Health, p.Hunger, p.Happiness, p.Energy, p.Hygiene, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) SavePet(ctx context.Context, p *entity.Pet, chatID int) error {
	_, err := r.db.Exec(ctx, sqlSavePet, chatID, p.Name, p.Health, p.Hunger, p.Happiness, p.Energy, p.Hygiene,
		p.State, p.SleepStartTime, p.Config.HungerDecayRate, p.Config.EnergyDecayRate, p.Config.HygieneDecayRate, p.Config.HappinessDecayRate, time.Now())

	return err
}

func (r *Repository) LoadPet(ctx context.Context, chatID int) (*entity.Pet, error) {
	var (
		p         entity.Pet
		petConfig entity.PetConfig
	)

	createdAt := time.Time{}
	err := r.db.QueryRow(ctx, sqlLoadPet, chatID).Scan(
		&p.Name, &p.Health, &p.Hunger, &p.Happiness, &p.Energy, &p.Hygiene,
		&p.State, &p.SleepStartTime, &petConfig.HungerDecayRate, &petConfig.EnergyDecayRate, &petConfig.HygieneDecayRate, &petConfig.HappinessDecayRate, &p.LastUpdated,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrPetNotFound
		}

		return nil, err
	}

	age := time.Since(createdAt).Hours() / 24

	p.Age = int(age)

	p.Config = petConfig

	return &p, nil
}

func (r *Repository) GetChats(ctx context.Context) ([]int, error) {
	rows, err := r.db.Query(ctx, sqlGetChats)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	chats := make([]int, 0)
	for rows.Next() {
		var chatId int

		err = rows.Scan(&chatId)
		if err != nil {
			return nil, err
		}

		chats = append(chats, chatId)
	}

	return chats, nil
}

func (r *Repository) GetLastAlert(ctx context.Context, chatID int, alertType string) (time.Time, error) {
	var lastAlert time.Time

	query := `SELECT last_alert FROM alerts WHERE chat_id = ? AND alert_type = ?`
	err := r.db.QueryRow(ctx, query, chatID, alertType).Scan(&lastAlert)

	// Если в БД еще нет записи, возвращаем старую дату (чтобы сразу отправить уведомление)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	} else if err != nil {
		return time.Time{}, err
	}

	return lastAlert, nil
}

func (r *Repository) UpdateLastAlert(ctx context.Context, chatID int, alertType string, now time.Time) error {
	query := `
	INSERT INTO alerts (chat_id, alert_type, last_alert) 
	VALUES (?, ?, ?) 
	ON CONFLICT(chat_id, alert_type) 
	DO UPDATE SET last_alert = excluded.last_alert;
	`
	_, err := r.db.Exec(ctx, query, chatID, alertType, now)

	return err
}
