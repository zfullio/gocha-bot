-- Таблица питомцев
CREATE TABLE IF NOT EXISTS pets.pets
(
    id                   SERIAL PRIMARY KEY,
    chat_id              BIGINT NOT NULL,
    name                 TEXT   NOT NULL,
    health               INTEGER   DEFAULT 100,
    hunger               INTEGER   DEFAULT 100,
    happiness            INTEGER   DEFAULT 100,
    energy               INTEGER   DEFAULT 100,
    hygiene              INTEGER   DEFAULT 100,
    state                TEXT      DEFAULT 'alive',
    sleep_start_time     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    hunger_decay_rate    INTEGER   DEFAULT 2,
    energy_decay_rate    INTEGER   DEFAULT 3,
    hygiene_decay_rate   INTEGER   DEFAULT 1,
    happiness_decay_rate INTEGER   DEFAULT 1,
    last_updated         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active            BOOL      DEFAULT TRUE
);

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS pets.users
(
    id      SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,  -- ID чата
    user_id BIGINT NOT NULL,  -- ID пользователя
    role    TEXT   NOT NULL,  -- Роль пользователя (например, "повар", "врач", "участник")
    UNIQUE (chat_id, user_id) -- Уникальный ключ для пары chat_id и user_id
);

-- Таблица оповещений
CREATE TABLE IF NOT EXISTS pets.alerts
(
    chat_id    BIGINT NOT NULL,
    alert_type TEXT   NOT NULL, -- 'health', 'hunger', 'happiness', 'energy', 'hygiene'
    last_alert TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chat_id, alert_type)
);