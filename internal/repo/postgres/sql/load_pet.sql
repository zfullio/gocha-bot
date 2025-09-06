SELECT
    name,
    health,
    hunger,
    happiness,
    energy,
    hygiene,
    state,
    sleep_start_time,
    hunger_decay_rate,
    energy_decay_rate,
    hygiene_decay_rate,
    happiness_decay_rate,
    last_updated,
    created_at
FROM pets.pets
WHERE chat_id = $1 and is_active = true;