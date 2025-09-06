UPDATE pets.pets
SET is_active = FALSE
WHERE chat_id = $1;