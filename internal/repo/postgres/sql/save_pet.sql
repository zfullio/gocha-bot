INSERT INTO pets.pets (
    chat_id,
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
    last_updated
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
         )
ON CONFLICT(chat_id)
    DO UPDATE SET
                  name                 = EXCLUDED.name,
                  health               = EXCLUDED.health,
                  hunger               = EXCLUDED.hunger,
                  happiness            = EXCLUDED.happiness,
                  energy               = EXCLUDED.energy,
                  hygiene              = EXCLUDED.hygiene,
                  state                = EXCLUDED.state,
                  sleep_start_time     = EXCLUDED.sleep_start_time,
                  hunger_decay_rate    = EXCLUDED.hunger_decay_rate,
                  energy_decay_rate    = EXCLUDED.energy_decay_rate,
                  hygiene_decay_rate   = EXCLUDED.hygiene_decay_rate,
                  happiness_decay_rate = EXCLUDED.happiness_decay_rate,
                  last_updated         = EXCLUDED.last_updated;