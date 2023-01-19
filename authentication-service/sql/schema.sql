CREATE TABLE users
(
    id         serial PRIMARY KEY,
    username   text NOT NULL UNIQUE,
    email      text NOT NULL UNIQUE,
    password   text NOT NULL,
    first_name text NOT NULL,
    last_name  text NOT NULL
);