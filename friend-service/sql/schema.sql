CREATE TABLE users
(
    id         serial PRIMARY KEY,
    user_id int NOT NULL UNIQUE,
    username   text NOT NULL UNIQUE,
    email      text NOT NULL UNIQUE,
    first_name text NOT NULL,
    last_name  text NOT NULL
);

CREATE TABLE follow
(
    id         serial PRIMARY KEY,
    friend_a int NOT NULL REFERENCES users(id),
    friend_b int NOT NULL REFERENCES users(id),
    UNIQUE (friend_a, friend_b)
);