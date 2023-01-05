CREATE TABLE users
(
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username   text NOT NULL UNIQUE,
    email      text NOT NULL UNIQUE,
    password   text NOT NULL,
    first_name text NOT NULL,
    last_name  text NOT NULL
);

CREATE TABLE user_profile_images(
                           user_id uuid NOT NULL UNIQUE,
                           image_id uuid NOT NULL,
                           PRIMARY KEY (user_id, image_id)
);