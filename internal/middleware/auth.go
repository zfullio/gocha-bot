package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// TelegramUser данные пользователя из initData.
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// VerifyTelegramWebAppData проверяет валидность initData от Telegram Web App.
func VerifyTelegramWebAppData(token, initData string) bool {
	if initData == "" {
		return false
	}

	dataPairs := strings.Split(initData, "&")
	sort.Strings(dataPairs)

	var (
		dataCheckString string
		receivedHash    string
	)

	for _, pair := range dataPairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		if key == "hash" {
			receivedHash = value
		} else {
			decodedValue, err := url.QueryUnescape(value)
			if err != nil {
				decodedValue = value
			}

			if dataCheckString != "" {
				dataCheckString += "\n"
			}

			dataCheckString += key + "=" + decodedValue
		}
	}

	if receivedHash == "" {
		return false
	}

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(token))

	computedHash := hmac.New(sha256.New, secretKey.Sum(nil))
	computedHash.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(computedHash.Sum(nil))

	return receivedHash == expectedHash
}

// ExtractUserFromInitData извлекает данные пользователя из initData.
func ExtractUserFromInitData(initData string) (*TelegramUser, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, err
	}

	userStr := values.Get("user")
	if userStr == "" {
		return nil, nil
	}

	var user TelegramUser
	if err := json.Unmarshal([]byte(userStr), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// TelegramAuthMiddleware создает middleware для проверки аутентификации Telegram.
func TelegramAuthMiddleware(botToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			initData := r.Header.Get("X-Telegram-Init-Data")
			fmt.Printf("tg-data-header: %s", initData)

			if initData == "" {
				// http.Error(w, `{"error": "X-Telegram-Init-Data header required"}`, http.StatusUnauthorized)
				return
			}

			if !VerifyTelegramWebAppData(botToken, initData) {
				http.Error(w, `{"error": "Invalid Telegram init data"}`, http.StatusUnauthorized)

				return
			}

			// Извлекаем пользователя и добавляем в контекст
			user, err := ExtractUserFromInitData(initData)
			if err != nil {
				http.Error(w, `{"error": "Failed to extract user data"}`, http.StatusUnauthorized)

				return
			}

			// Добавляем пользователя в контекст
			ctx := context.WithValue(r.Context(), "telegram_user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext извлекает пользователя из контекста.
func GetUserFromContext(r *http.Request) (*TelegramUser, bool) {
	user, ok := r.Context().Value("telegram_user").(*TelegramUser)

	return user, ok
}
