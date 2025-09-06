// Получаем API URL
const API_BASE_URL = window.API_BASE_URL || 'http://localhost:8080';

// === ИНИЦИАЛИЗАЦИЯ ===
let tg = null;
let petData = null;
let isLoading = false;

// Дожидаемся полной загрузки страницы
document.addEventListener('DOMContentLoaded', async () => {
    console.log('DOM loaded, initializing...');

    try {
        if (typeof window.Telegram !== 'undefined' && window.Telegram.WebApp && window.Telegram.WebApp.initData) {
            // настоящее окружение Телеграма
            tg = window.Telegram.WebApp;
            console.log('✅ Telegram WebApp detected:', tg.initData);
        } else if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
            // режим разработки с моками
            console.log('⚙️ Localhost detected, using mock');
            // Создаем мок для разработки
            try {
                const res = await fetch(`${API_BASE_URL}/api/debug/init-config`);
                const config = await res.json();

                // Восстанавливаем функции из строк
                function reviveMethods(obj) {
                    for (const key in obj) {
                        if (typeof obj[key] === "string" && obj[key].startsWith("function")) {
                            try {
                                obj[key] = eval(`(${obj[key]})`);
                            } catch (e) {
                                console.warn(`Failed to revive method ${key}:`, e);
                                obj[key] = function () {
                                    console.log(`Mock function ${key} called`);
                                };
                            }
                        } else if (typeof obj[key] === "object" && obj[key] !== null) {
                            reviveMethods(obj[key]);
                        }
                    }
                }

                reviveMethods(config);

                window.Telegram = {
                    WebApp: {
                        ...config,
                        MainButton: config.mainButton || {
                            setText: function (text) {
                                this.text = text;
                            },
                            show: function () {
                                this.isVisible = true;
                            },
                            hide: function () {
                                this.isVisible = false;
                            },
                            onClick: function (callback) {
                                this.clickCallback = callback;
                            },
                            offClick: function (callback) {
                                this.clickCallback = null;
                            },
                            showProgress: function () {
                                console.log('Showing progress');
                            },
                            hideProgress: function () {
                                console.log('Hiding progress');
                            }
                        },
                        HapticFeedback: config.hapticFeedback || {
                            impactOccurred: function (type) {
                                console.log(`Haptic impact: ${type}`);
                            },
                            notificationOccurred: function (type) {
                                console.log(`Haptic notification: ${type}`);
                            }
                        },
                        onEvent: function (event, callback) {
                            console.log(`Event listener: ${event}`);
                        },
                        expand: function () {
                            console.log('WebApp expanded');
                        },
                        ready: function () {
                            console.log('WebApp ready');
                        },
                        ...config.methods,
                    }
                };

                tg = window.Telegram.WebApp;
                console.log('Mock Telegram WebApp created with initData:', tg.initData);
            } catch (e) {
                console.error('Failed to load debug config:', e);
                // Создаем минимальный мок
                tg = createMinimalMock();
            }
        } else {
            console.error('❌ Not in Telegram environment and not on localhost');
            showNotification('Приложение работает только в Telegram или на localhost для разработки.', 'danger');
        }

        // === Проверка initData ===
        if (!tg || !tg.initData) {
            console.error('initData is missing');
            showNotification('Ошибка: отсутствуют данные авторизации Telegram', 'danger');
            return;
        }

        // Инициализация
        window.tg = tg;

        // Безопасный вызов методов WebApp
        try {
            if (typeof tg.expand === 'function') tg.expand();
            if (typeof tg.ready === 'function') tg.ready();
        } catch (e) {
            console.warn('Error calling WebApp methods:', e);
        }

        updateThemeColors();
        updateViewportHeight();

        // Безопасная установка обработчиков событий
        try {
            if (typeof tg.onEvent === 'function') {
                tg.onEvent('themeChanged', updateThemeColors);
                tg.onEvent('viewportChanged', updateViewportHeight);
            }
        } catch (e) {
            console.warn('Error setting event listeners:', e);
        }

        await loadPetInfo();
    } catch (error) {
        console.error('Initialization failed:', error);
        showNotification(`Ошибка инициализации: ${error.message}`, 'danger');
    }
});

// Создание минимального мока для случаев когда нет доступа к серверу
function createMinimalMock() {
    const fallbackInitData = 'user=%7B%22id%22%3A12345%2C%22first_name%22%3A%22Test%22%2C%22last_name%22%3A%22User%22%2C%22username%22%3A%22testuser%22%7D&auth_date=1640995200&hash=test_hash';

    return {
        initData: fallbackInitData,
        themeParams: {
            bg_color: '#ffffff',
            text_color: '#000000',
            button_color: '#6366f1'
        },
        viewportHeight: window.innerHeight,
        MainButton: {
            text: '',
            isVisible: false,
            setText: function (text) {
                this.text = text;
                console.log('MainButton text:', text);
            },
            show: function () {
                this.isVisible = true;
                console.log('MainButton shown');
            },
            hide: function () {
                this.isVisible = false;
                console.log('MainButton hidden');
            },
            onClick: function (callback) {
                this.clickCallback = callback;
                console.log('MainButton onClick set');
            },
            offClick: function (callback) {
                this.clickCallback = null;
                console.log('MainButton onClick removed');
            },
            showProgress: function () {
                console.log('Showing progress');
            },
            hideProgress: function () {
                console.log('Hiding progress');
            }
        },
        HapticFeedback: {
            impactOccurred: function (type) {
                console.log(`Haptic impact: ${type}`);
            },
            notificationOccurred: function (type) {
                console.log(`Haptic notification: ${type}`);
            }
        },
        onEvent: function (event, callback) {
            console.log(`Event listener registered: ${event}`);
        },
        expand: function () {
            console.log('WebApp expanded');
        },
        ready: function () {
            console.log('WebApp ready');
        }
    };
}

// Обновление цветов и высоты
function updateThemeColors() {
    if (!tg || !tg.themeParams) return;

    const root = document.documentElement;
    root.style.setProperty('--tg-theme-bg-color', tg.themeParams.bg_color || '#ffffff');
    root.style.setProperty('--tg-theme-text-color', tg.themeParams.text_color || '#000000');
    root.style.setProperty('--tg-theme-button-color', tg.themeParams.button_color || '#6366f1');
}

function updateViewportHeight() {
    if (!tg) return;

    const height = tg.viewportHeight || window.innerHeight;
    document.documentElement.style.setProperty('--tg-viewport-height', `${height}px`);
}

// Загрузка информации о питомце
async function loadPetInfo() {
    if (isLoading || !tg) return;

    console.log('loadPetInfo: initData =', tg.initData);

    if (!tg.initData || tg.initData === '') {
        console.error('InitData is still empty, cannot make API request');
        showNotification('Ошибка: нет данных авторизации', 'danger');
        return;
    }

    isLoading = true;
    showLoading(true);

    try {
        const requestUrl = `${API_BASE_URL}/api/pet/info/`;
        const requestHeaders = {
            'Content-Type': 'application/json',
            'X-Telegram-Init-Data': tg.initData
        };

        console.log('Making request to:', requestUrl);

        const response = await fetch(requestUrl, {
            method: 'GET',
            headers: requestHeaders,
            mode: 'cors'
        });

        console.log('API Response status:', response.status);

        if (!response.ok) {
            let errorMessage = 'Ошибка загрузки питомца';
            try {
                const errorData = await response.json();
                console.log('Error response:', errorData);

                // Проверяем унифицированный формат ответа
                if (errorData.message === "Питомец не найден") {
                    console.log('Pet not found, showing create screen');
                    showPetNotFound();
                    return;
                }

                errorMessage = errorData.message || errorMessage;
            } catch (e) {
                errorMessage += ` (${response.status}: ${response.statusText})`;
            }
            throw new Error(errorMessage);
        }

        const apiResponse = await response.json();
        console.log('API Response:', apiResponse);

        // Проверяем унифицированный формат ответа
        if (!apiResponse.success) {
            if (apiResponse.message === "Питомец не найден") {
                console.log('Pet not found, showing create screen');
                showPetNotFound();
                return;
            }
            throw new Error(apiResponse.message || 'Ошибка загрузки питомца');
        }

        // Данные питомца теперь в поле data
        petData = apiResponse.data;
        console.log('Pet data loaded:', petData);
        displayPetInfo();

        // Безопасный вызов HapticFeedback
        if (tg.HapticFeedback && typeof tg.HapticFeedback.notificationOccurred === 'function') {
            tg.HapticFeedback.notificationOccurred('success');
        }
    } catch (error) {
        console.error('Ошибка загрузки питомца:', error);
        showNotification(error.message || 'Не удалось загрузить питомца', 'danger');
    } finally {
        showLoading(false);
        isLoading = false;
    }
}

// Показать экран "питомец не найден"
function showPetNotFound() {
    console.log('Showing pet not found screen');

    const petInfoEl = document.getElementById('petInfo');
    const createPetScreenEl = document.getElementById('createPetScreen');

    if (petInfoEl) petInfoEl.style.display = 'none';
    if (createPetScreenEl) createPetScreenEl.style.display = 'block';

    const header = document.querySelector('header');
    if (header) {
        header.style.display = 'block';
    }

    if (tg && tg.MainButton) {
        tg.MainButton.setText("Создать питомца");
        tg.MainButton.color = "#6366f1";
        tg.MainButton.show();

        // Безопасное удаление и добавление обработчика
        try {
            if (typeof tg.MainButton.offClick === 'function') {
                tg.MainButton.offClick(handleMainButtonClick);
            }
            if (typeof tg.MainButton.onClick === 'function') {
                tg.MainButton.onClick(handleMainButtonClick);
            }
        } catch (e) {
            console.warn('Error setting MainButton handlers:', e);
        }
    }
}

// Показать экран создания нового питомца (для мертвого питомца)
function showCreateNewPetScreen() {
    console.log('Showing create new pet screen');
    petData = null; // Сбрасываем данные мертвого питомца

    const petInfoEl = document.getElementById('petInfo');
    const createPetScreenEl = document.getElementById('createPetScreen');

    if (petInfoEl) petInfoEl.style.display = 'none';
    if (createPetScreenEl) createPetScreenEl.style.display = 'block';

    const header = document.querySelector('header');
    if (header) {
        header.style.display = 'block';
    }

    // Очищаем поле ввода имени
    const petNameInput = document.getElementById('petNameInput');
    if (petNameInput) {
        petNameInput.value = '';
        try {
            petNameInput.focus();
        } catch (e) {
            console.warn('Could not focus input:', e);
        }
    }

    if (tg && tg.MainButton) {
        tg.MainButton.setText("Создать питомца");
        tg.MainButton.color = "#6366f1";
        tg.MainButton.show();

        try {
            if (typeof tg.MainButton.offClick === 'function') {
                tg.MainButton.offClick(handleMainButtonClick);
            }
            if (typeof tg.MainButton.onClick === 'function') {
                tg.MainButton.onClick(handleMainButtonClick);
            }
        } catch (e) {
            console.warn('Error setting MainButton handlers:', e);
        }
    }
}

// Главная кнопка: создать питомца
function handleMainButtonClick() {
    console.log('Main button clicked, petData:', !!petData);
    if (!petData) {
        createPet();
    }
}

// Отобразить питомца
function displayPetInfo() {
    console.log('Displaying pet info');

    const createPetScreenEl = document.getElementById('createPetScreen');
    const petInfoEl = document.getElementById('petInfo');

    if (createPetScreenEl) createPetScreenEl.style.display = 'none';
    if (petInfoEl) petInfoEl.style.display = 'block';

    const header = document.querySelector('header');
    if (header) {
        header.style.display = 'none';
    }

    // Скрываем главную кнопку когда показываем питомца
    if (tg && tg.MainButton && typeof tg.MainButton.hide === 'function') {
        tg.MainButton.hide();
    }

    updatePetDisplay();
}

// Создать нового питомца
async function createPet() {
    if (isLoading || !tg) return;

    console.log('createPet: initData =', tg.initData);

    if (!tg.initData || tg.initData === '') {
        console.error('InitData is empty, cannot create pet');
        showNotification('Ошибка: нет данных авторизации', 'danger');
        return;
    }

    // Получаем имя питомца из поля ввода
    const petNameInput = document.getElementById('petNameInput');
    const petName = petNameInput ? petNameInput.value.trim() : '';

    if (!petName) {
        showNotification('Введите имя для питомца', 'warning');
        if (petNameInput) {
            try {
                petNameInput.focus();
            } catch (e) {
                console.warn('Could not focus input:', e);
            }
        }
        return;
    }

    isLoading = true;
    showLoading(true);

    try {
        const requestUrl = `${API_BASE_URL}/api/pet/create/`;
        const requestHeaders = {
            'Content-Type': 'application/json',
            'X-Telegram-Init-Data': tg.initData
        };

        const requestBody = JSON.stringify({
            name: petName
        });

        const response = await fetch(requestUrl, {
            method: 'POST',
            headers: requestHeaders,
            body: requestBody,
            mode: 'cors'
        });

        if (!response.ok) {
            let errorMessage = 'Не удалось создать питомца';
            try {
                const errorData = await response.json();
                console.log('Error response:', errorData);
                errorMessage = errorData.message || errorMessage;
            } catch (e) {
                console.log('Could not read error body:', e);
                errorMessage += ` (${response.status}: ${response.statusText})`;
            }
            throw new Error(errorMessage);
        }

        const apiResponse = await response.json();
        console.log('Create pet response:', apiResponse);

        // Проверяем унифицированный формат ответа
        if (!apiResponse.success) {
            throw new Error(apiResponse.message || 'Не удалось создать питомца');
        }

        // Показываем сообщение об успешном создании
        showNotification(apiResponse.message || `🎉 Поздравляем! Питомец ${petName} создан!`, 'good');

        if (tg.HapticFeedback && typeof tg.HapticFeedback.notificationOccurred === 'function') {
            tg.HapticFeedback.notificationOccurred('success');
        }

        // После создания загружаем информацию о питомце
        await loadPetInfo();
    } catch (error) {
        console.error('Ошибка создания питомца:', error);
        showNotification(error.message || 'Не удалось создать питомца', 'danger');

        if (tg.HapticFeedback && typeof tg.HapticFeedback.notificationOccurred === 'function') {
            tg.HapticFeedback.notificationOccurred('error');
        }
    } finally {
        showLoading(false);
        isLoading = false;
    }
}

// Выполнить действие (кормить, играть и т.д.)
async function performAction(action) {
    if (isLoading || !tg) return;

    if (!petData) {
        console.warn('petData is empty, reloading...');
        await loadPetInfo();
        if (!petData) {
            showNotification('Ошибка: питомец не загружен', 'danger');
            return;
        }
    }

    console.log('performAction: initData =', tg.initData);

    if (!tg.initData || tg.initData === '') {
        console.error('InitData is empty, cannot perform action');
        showNotification('Ошибка: нет данных авторизации', 'danger');
        return;
    }

    // Проверяем доступность действия на фронтенде (для UX)
    const availableActions = petData.availableActions;
    if (availableActions) {
        const actionMap = {
            'feed': 'canFeed',
            'play': 'canPlay',
            'clean': 'canClean',
            'heal': 'canHeal',
            'sleep': 'canSleep',
            'wakeup': 'canWakeUp'
        };

        if (!availableActions[actionMap[action]]) {
            showNotification(getActionDisabledReason(action, availableActions), 'warning');
            return;
        }
    }

    // Безопасный вызов HapticFeedback
    if (tg.HapticFeedback && typeof tg.HapticFeedback.impactOccurred === 'function') {
        tg.HapticFeedback.impactOccurred('medium');
    }

    isLoading = true;
    showLoading(true);
    setActionsEnabled(false);

    try {
        const response = await fetch(`${API_BASE_URL}/api/pet/${action}/`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': tg.initData
            },
            mode: 'cors'
        });

        if (!response.ok) {
            let errorMessage = 'Ошибка действия';
            try {
                const errorData = await response.json();
                console.log('Action error response:', errorData);

                if (errorData.message === "Питомец не найден") {
                    showPetNotFound();
                    return;
                }

                errorMessage = errorData.message || errorMessage;
            } catch (e) {
                errorMessage += ` (${response.status}: ${response.statusText})`;
            }
            throw new Error(errorMessage);
        }

        const apiResponse = await response.json();
        console.log('Action response:', apiResponse);

        if (!apiResponse.success) {
            if (apiResponse.message === "Питомец не найден") {
                showPetNotFound();
                return;
            }
            throw new Error(apiResponse.message || 'Ошибка при выполнении действия');
        }

        // Обновляем данные питомца из ответа
        const actionResult = apiResponse.data; // PetActionResult
        petData = actionResult.pet;

        updatePetDisplay();

        // Показываем обратную связь с бэкенда
        if (actionResult.actionFeedback) {
            showNotification(actionResult.actionFeedback, 'good');
        }

        if (tg.HapticFeedback && typeof tg.HapticFeedback.notificationOccurred === 'function') {
            tg.HapticFeedback.notificationOccurred('success');
        }
    } catch (error) {
        console.error(`Ошибка действия ${action}:`, error);
        showNotification(error.message || 'Не удалось выполнить действие', 'danger');

        if (tg.HapticFeedback && typeof tg.HapticFeedback.notificationOccurred === 'function') {
            tg.HapticFeedback.notificationOccurred('error');
        }
    } finally {
        showLoading(false);
        setActionsEnabled(true);
        isLoading = false;
    }
}

// Включить/отключить кнопки действий
function setActionsEnabled(enabled) {
    document.querySelectorAll('.action-btn').forEach(btn => {
        btn.disabled = !enabled;
        btn.style.opacity = enabled ? '1' : '0.5';

        if (enabled && tg && tg.HapticFeedback) {
            // Удаляем старые обработчики перед добавлением новых
            btn.removeEventListener('touchstart', handleTouchStart);
            btn.addEventListener('touchstart', handleTouchStart, {passive: true});
        }
    });
}

// Отдельная функция для обработки тач событий
function handleTouchStart() {
    if (tg && tg.HapticFeedback && typeof tg.HapticFeedback.impactOccurred === 'function') {
        tg.HapticFeedback.impactOccurred('light');
    }
}

// Упрощенная функция обновления интерфейса питомца
function updatePetDisplay() {
    if (!petData) return;

    const pet = petData;
    const stats = {
        hunger: pet.hunger || 0,
        happiness: pet.happiness || 0,
        hygiene: pet.hygiene || 0,
        health: pet.health || 0,
        energy: pet.energy || 0
    };

    // === Используем статус с бэкенда ===
    const status = pet.status || {};
    const uiConfig = pet.uiConfig || {criticalThreshold: 20, warningThreshold: 40};

    // === Обновляем статы с учетом конфигурации ===
    Object.keys(stats).forEach(stat => {
        const valueEl = document.getElementById(`${stat}Value`);
        const barEl = document.getElementById(`${stat}Bar`);

        if (valueEl) {
            animateNumberChange(valueEl, stats[stat]);
            // Используем пороги с бэкенда
            if (stats[stat] <= uiConfig.criticalThreshold) {
                valueEl.style.color = '#ef4444';
            } else if (stats[stat] <= uiConfig.warningThreshold) {
                valueEl.style.color = '#f59e0b';
            } else {
                valueEl.style.color = 'var(--text)';
            }
        }

        if (barEl) {
            const pct = Math.max(0, Math.min(100, stats[stat]));
            barEl.style.width = `${pct}%`;
            if (pct <= uiConfig.criticalThreshold) {
                barEl.style.background = 'linear-gradient(90deg, #ef4444, #f87171)';
            } else if (pct <= uiConfig.warningThreshold) {
                barEl.style.background = 'linear-gradient(90deg, #f59e0b, #fbbf24)';
            } else {
                barEl.style.background = 'linear-gradient(90deg, #6366f1, #8b5cf6)';
            }
        }
    });

    // === Имя питомца ===
    const nameEl = document.getElementById('petName');
    if (nameEl) nameEl.textContent = pet.name || "Безымянный";

    // === Аватар из бэкенда ===
    updateAvatarFromBackend(pet.avatar);

    // === Статусное сообщение с бэкенда ===
    updateStatusFromBackend(status);

    // === Эффекты при критическом состоянии ===
    updateCriticalEffectsFromBackend(status);

    // === Обновление доступных действий ===
    updateAvailableActions(pet.availableActions);

    // === Интерфейс для мертвого питомца ===
    updateDeadPetInterface(pet.state === "dead");

    // === Обновление заголовка ===
    document.title = `${pet.name || "Питомец"} - Тамагочи`;
}

// Обновление аватара из бэкенда
function updateAvatarFromBackend(avatar) {
    const moodImg = document.getElementById('moodImage');
    const emojiAvatar = document.getElementById('emojiAvatar');
    const moodIndicator = document.getElementById('moodIndicator');

    if (!moodImg || !emojiAvatar || !moodIndicator) return;

    const avatarData = avatar || {};
    const imgSrc = avatarData.image;
    const emoji = avatarData.emoji || '🐱';
    const moodEmoji = avatarData.mood || '😐';

    if (imgSrc) {
        moodImg.src = imgSrc;
        moodImg.style.display = 'block';
        emojiAvatar.style.display = 'none';

        // Добавляем обработчик ошибки загрузки изображения
        moodImg.onerror = function () {
            console.warn('Failed to load pet image:', imgSrc);
            moodImg.style.display = 'none';
            emojiAvatar.style.display = 'block';
            emojiAvatar.textContent = emoji;
        };
    } else {
        moodImg.style.display = 'none';
        emojiAvatar.style.display = 'block';
        emojiAvatar.textContent = emoji;
    }

    moodIndicator.textContent = moodEmoji;
}

// Обновление статуса из бэкенда
function updateStatusFromBackend(status) {
    const statusEl = document.getElementById('statusMessage');
    if (!statusEl) return;

    const message = status.statusMessage || 'Всё в порядке';
    const type = status.statusType || 'good';

    statusEl.textContent = message;
    statusEl.className = `status-message status-${type}`;
}

// Эффекты при критическом состоянии из бэкенда
function updateCriticalEffectsFromBackend(status) {
    const container = document.getElementById('petContainer');
    if (!container) return;

    if (status.isCritical) {
        container.classList.add('critical-pulse');
    } else {
        container.classList.remove('critical-pulse');
    }
}

// Обновление доступности действий
function updateAvailableActions(availableActions) {
    if (!availableActions) return;

    const actionMap = {
        'feed': 'canFeed',
        'play': 'canPlay',
        'clean': 'canClean',
        'heal': 'canHeal',
        'sleep': 'canSleep',
        'wakeup': 'canWakeUp'
    };

    Object.keys(actionMap).forEach(action => {
        const btn = document.getElementById(`${action}Btn`);
        if (btn) {
            const canPerform = availableActions[actionMap[action]];
            btn.disabled = !canPerform;
            btn.style.opacity = canPerform ? '1' : '0.5';

            // Добавляем подсказку почему действие недоступно
            if (!canPerform) {
                btn.title = getActionDisabledReason(action, availableActions);
            } else {
                btn.title = '';
            }
        }
    });

    // Обновляем кнопки сна/пробуждения
    const sleepBtn = document.getElementById('sleepBtn');
    const wakeupBtn = document.getElementById('wakeupBtn');

    if (sleepBtn && wakeupBtn) {
        if (availableActions.canWakeUp) {
            sleepBtn.style.display = 'none';
            wakeupBtn.style.display = 'flex';
        } else {
            sleepBtn.style.display = 'flex';
            wakeupBtn.style.display = 'none';
        }
    }
}

// Получение причины недоступности действия
function getActionDisabledReason(action, availableActions) {
    // Эти сообщения тоже можно получать с бэкенда в будущем
    const reasons = {
        'feed': !availableActions.canFeed ? 'Питомец не голоден или спит' : '',
        'play': !availableActions.canPlay ? 'Питомец устал или уже счастлив' : '',
        'clean': !availableActions.canClean ? 'Питомец уже чистый или спит' : '',
        'heal': !availableActions.canHeal ? 'Питомец здоров или спит' : '',
        'sleep': !availableActions.canSleep ? 'Питомец не устал или уже спит' : '',
        'wakeup': !availableActions.canWakeUp ? 'Питомец не спит' : ''
    };

    return reasons[action] || '';
}

// Обновление интерфейса для мертвого питомца
function updateDeadPetInterface(isDead) {
    const actionsGrid = document.querySelector('.actions-grid');
    let createNewBtn = document.getElementById('createNewPetBtn');

    if (isDead) {
        // Скрываем кнопки действий
        if (actionsGrid) {
            actionsGrid.style.display = 'none';
        }

        // Показываем кнопку создания нового питомца
        if (!createNewBtn) {
            createNewBtn = document.createElement('button');
            createNewBtn.id = 'createNewPetBtn';
            createNewBtn.className = 'create-new-pet-btn';
            createNewBtn.textContent = '💫 Создать нового питомца';
            createNewBtn.onclick = showCreateNewPetScreen;

            const statusMessage = document.getElementById('statusMessage');
            if (statusMessage && statusMessage.parentNode) {
                statusMessage.parentNode.insertBefore(createNewBtn, statusMessage.nextSibling);
            }
        }
        if (createNewBtn) createNewBtn.style.display = 'block';
    } else {
        // Показываем кнопки действий
        if (actionsGrid) {
            actionsGrid.style.display = 'grid';
        }

        // Скрываем кнопку создания нового питомца
        if (createNewBtn) {
            createNewBtn.style.display = 'none';
        }
    }
}

// Анимация изменения числа
function animateNumberChange(el, newValue) {
    if (!el) return;

    const currentValue = parseInt(el.textContent) || 0;
    if (currentValue === newValue) return;

    const steps = 10;
    const increment = (newValue - currentValue) / steps;
    let step = 0;

    // Очищаем предыдущую анимацию если она есть
    if (el.animationTimer) {
        clearInterval(el.animationTimer);
    }

    el.animationTimer = setInterval(() => {
        step++;
        const value = Math.round(currentValue + increment * step);
        el.textContent = value;

        if (step === steps) {
            clearInterval(el.animationTimer);
            el.textContent = newValue;
            el.animationTimer = null;
        }
    }, 50);
}

// Показать/скрыть лоадер
function showLoading(show) {
    const loader = document.getElementById('loadingSpinner');
    if (loader) {
        loader.style.display = show ? 'block' : 'none';
    }

    if (tg && tg.MainButton) {
        try {
            if (show && typeof tg.MainButton.showProgress === 'function') {
                tg.MainButton.showProgress();
            } else if (!show && typeof tg.MainButton.hideProgress === 'function') {
                tg.MainButton.hideProgress();
            }
        } catch (e) {
            console.warn('Error controlling MainButton progress:', e);
        }
    }
}

// Уведомления
function showNotification(message, type = 'good') {
    // Создаем временное уведомление если нет statusMessage
    let el = document.getElementById('statusMessage');

    if (!el) {
        // Создаем временный элемент для уведомления
        el = document.createElement('div');
        el.id = 'tempNotification';
        el.className = 'status-message';
        el.style.cssText = `
            position: fixed;
            top: 20px;
            left: 50%;
            transform: translateX(-50%);
            z-index: 1000;
            padding: 10px 20px;
            border-radius: 8px;
            font-weight: 500;
            text-align: center;
            max-width: 90%;
            word-wrap: break-word;
        `;
        document.body.appendChild(el);

        // Удаляем временный элемент через 5 секунд
        setTimeout(() => {
            if (el && el.id === 'tempNotification' && el.parentNode) {
                el.parentNode.removeChild(el);
            }
        }, 5000);
    }

    if (!el) return;

    el.textContent = message;
    el.className = `status-message status-${type}`;

    // Если это основной statusMessage, восстанавливаем его через 3 секунды
    if (el.id === 'statusMessage') {
        setTimeout(() => {
            if (petData && el.parentNode) {
                updatePetDisplay();
            }
        }, 3000);
    }
}

// Упрощенная функция отображения обратной связи
function showActionFeedback(action, result) {
    // Теперь сообщение приходит готовым с бэкенда
    const message = result.actionFeedback || result.message || 'Действие выполнено';
    const type = result.success ? 'good' : 'danger';
    showNotification(message, type);
}

// Автообновление (только если страница активна)
let autoUpdateInterval = null;

function startAutoUpdate() {
    if (autoUpdateInterval) return; // Уже запущено

    autoUpdateInterval = setInterval(() => {
        if (petData && !isLoading && !document.hidden && tg) {
            loadPetInfo().catch(error => {
                console.warn('Auto-update failed:', error);
                // Не показываем ошибку пользователю для автообновления
            });
        }
    }, 30000);
}

function stopAutoUpdate() {
    if (autoUpdateInterval) {
        clearInterval(autoUpdateInterval);
        autoUpdateInterval = null;
    }
}

// Запускаем автообновление после успешной загрузки
document.addEventListener('DOMContentLoaded', () => {
    setTimeout(() => {
        if (petData) {
            startAutoUpdate();
        }
    }, 5000); // Задержка 5 секунд после загрузки
});

// Останавливаем автообновление когда страница скрыта
document.addEventListener('visibilitychange', () => {
    if (document.hidden) {
        stopAutoUpdate();
    } else {
        startAutoUpdate();
    }
});

// Обработка ошибок
window.addEventListener('unhandledrejection', e => {
    console.error('Unhandled rejection:', e.reason);
    showNotification('Произошла ошибка', 'danger');
    e.preventDefault(); // Предотвращаем показ ошибки в консоли браузера
});

window.addEventListener('error', e => {
    console.error('JavaScript error:', e.error);
    showNotification('Произошла ошибка', 'danger');
});

// Обработка состояния сети
window.addEventListener('online', () => {
    showNotification('Сеть восстановлена', 'good');
    // Перезагружаем данные при восстановлении сети
    if (petData && !isLoading) {
        loadPetInfo().catch(error => {
            console.warn('Failed to reload after network restore:', error);
        });
    }
});

window.addEventListener('offline', () => {
    showNotification('Нет подключения к сети', 'warning');
    stopAutoUpdate(); // Останавливаем автообновление при потере сети
});

// Дополнительная защита от множественных вызовов
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Дебаунс для критичных функций
const debouncedLoadPetInfo = debounce(loadPetInfo, 1000);
const debouncedCreatePet = debounce(createPet, 1000);

// Экспортируем основные функции для возможного использования в консоли разработчика
if (typeof window !== 'undefined') {
    window.tamagochi = {
        loadPetInfo: debouncedLoadPetInfo,
        createPet: debouncedCreatePet,
        performAction,
        showNotification,
        petData: () => petData,
        isLoading: () => isLoading
    };
}