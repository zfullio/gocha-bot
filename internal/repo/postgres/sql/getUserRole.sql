SELECT role
FROM pets.users
WHERE chat_id = $1 AND user_id = $2;