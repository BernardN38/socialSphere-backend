CREATE TABLE user_images(
                     id         serial PRIMARY KEY,
                    user_id int NOT NULL,
                    image_id uuid NOT NULL,
                    created_at timestamp NOT NULL DEFAULT NOW()
);
