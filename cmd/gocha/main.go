package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"

	"gocha/internal/config"
	"gocha/internal/handlers"
	"gocha/internal/repo/postgres"
	"gocha/internal/service"
	"gocha/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"github.com/rs/zerolog"
)

//go:embed ui/static/*
var staticFiles embed.FS

//go:embed ui/static/templates/*
var templateFS embed.FS

func main() {
	ctx := context.Background()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("can`t load configuration: %s", err)
	}

	var baseLogger zerolog.Logger

	var loggerCloser io.WriteCloser

	baseLogger, loggerCloser, err = logger.NewLogger(os.Stdout, cfg.Log.Level)

	coreLogger := logger.NewComponentLogger(baseLogger, "core", 2)
	repoLogger := logger.NewComponentLogger(baseLogger, "repo", 2)
	srvLogger := logger.NewComponentLogger(baseLogger, "service", 2)
	handlersLogger := logger.NewComponentLogger(baseLogger, "handlers", 2)

	defer func() {
		if loggerCloser != nil {
			err = loggerCloser.Close()
			if err != nil {
				log.Fatalf("error acquired while closing log writer: %+v", err)
			}
		}
	}()

	coreLogger.Info().Msg("application started")

	bot, err := telego.NewBot(cfg.TgToken, telego.WithDefaultLogger(false, true))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	botUser, err := bot.GetMe(ctx)
	if err != nil {
		coreLogger.Warn().Err(err).Msg("can't get bot info")
	}

	coreLogger.Info().Msgf("Bot user: %+v\n", botUser)

	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)

	bh, _ := th.NewBotHandler(bot, updates)

	initHandlers(bh)

	// Stop handling updates
	defer func() { _ = bh.Stop() }()

	// Start handling updates
	go func() {
		// Start handling updates
		err := bh.Start()
		if err != nil {
			coreLogger.Error().Err(err).Msg("bot handler failed")
		}
	}()

	mux := http.NewServeMux()

	// Создаём под-файл-систему, чтобы убрать префикс `ui/static` из путей
	subFS, err := fs.Sub(staticFiles, "ui/static")
	if err != nil {
		log.Fatal("Не удалось создать под-файл-систему:", err)
	}

	corsHandler := enableCORS(mux)

	pgPool, err := pgxpool.New(ctx, cfg.DbDataSourceName)
	if err != nil {
		coreLogger.Fatal().Stack().Err(err).Msg("Unable to create connection pool")
	}

	repo := postgres.NewRepository(&repoLogger, pgPool)

	srv := service.NewService(cfg, &srvLogger, repo)

	err = srv.MonitorPetsAll(ctx)
	if err != nil {
		log.Fatal("Не удалось запустить мониторинг питомцев")
	}

	defer srv.Stop()

	petHandlers := handlers.NewPetHandlers(handlersLogger, srv, cfg.BaseUrl, cfg.IsDev)

	if cfg.IsDev {
		mux.HandleFunc("/api/debug/mock-init-data", petHandlers.DebugMockInitDataHandler)
		mux.HandleFunc("/api/debug/init-config", petHandlers.DebugInitTgConfigHandler)
	}

	mux.HandleFunc("/api/pet/create/", petHandlers.PetNewHandler)
	mux.HandleFunc("/api/pet/info/", petHandlers.PetInfoHandler)
	mux.HandleFunc("/api/pet/heal/", petHandlers.PetHealHandler)
	mux.HandleFunc("/api/pet/feed/", petHandlers.PetFeedHandler)
	mux.HandleFunc("/api/pet/play/", petHandlers.PetPlayHandler)
	mux.HandleFunc("/api/pet/clean/", petHandlers.PetCleanHandler)
	mux.HandleFunc("/api/pet/sleep/", petHandlers.PetSleepHandler)
	mux.HandleFunc("/api/pet/wakeup/", petHandlers.PetWakeUpHandler)

	fileServer := http.FileServer(http.FS(subFS))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)

			return
		}

		tmpl, err := template.ParseFS(templateFS, "ui/static/templates/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		// Передаем API URL в шаблон
		data := map[string]interface{}{
			"APIBaseURL": cfg.BaseUrl,
		}

		tmpl.Execute(w, data)
	})

	coreLogger.Info().Msgf("Addr %s:%v", cfg.Host, cfg.Port)

	err = http.ListenAndServe(fmt.Sprintf("%s:%v", cfg.Host, cfg.Port), corsHandler)
	if err != nil {
		log.Fatal(err)
	}
}

func initHandlers(bh *th.BotHandler) {
	handlers.RunApp(bh)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы с фронтенда
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:63342")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Обрабатываем preflight запросы
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)

			return
		}

		next.ServeHTTP(w, r)
	})
}
