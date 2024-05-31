CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL
);

CREATE TABLE memecoins (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    links TEXT,
    descriptions TEXT,
    points INT NOT NULL,
    memecoin_id INT references memecoins(id)
);

CREATE TABLE user_tasks (
    user_id INT REFERENCES users(id),
    task_id INT REFERENCES tasks(id),
    completed BOOLEAN DEFAULT FALSE,
    proof_file_url TEXT NOT NULL,
    PRIMARY KEY (user_id, task_id)
);
