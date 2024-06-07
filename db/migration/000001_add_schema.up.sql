CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL
);

CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    name TEXT,
    links TEXT,
    descriptions TEXT
);

CREATE TABLE sub_tasks (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    links TEXT,
    description TEXT,
    points INT NOT NULL,
    task_id INT REFERENCES tasks(id)
);

CREATE TABLE user_sub_tasks (
    user_id INT REFERENCES users(id),
    sub_task_id INT REFERENCES sub_tasks(id),
    completed BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (user_id, sub_task_id)
);
