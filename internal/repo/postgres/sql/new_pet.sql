INSERT INTO pets.pets (
    chat_id,
    name,
    health,
    hunger,
    happiness,
    energy,
    hygiene,
    last_updated
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);