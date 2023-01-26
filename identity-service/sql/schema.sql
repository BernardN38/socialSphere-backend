CREATE TABLE users
(
    id        serial PRIMARY KEY,
    username   text NOT NULL UNIQUE,
    email      text NOT NULL UNIQUE,
    first_name text NOT NULL,
    last_name  text NOT NULL
);

CREATE TABLE user_profile_images(
                           user_id int NOT NULL UNIQUE,
                           image_id uuid NOT NULL,
                           PRIMARY KEY (user_id, image_id)
);