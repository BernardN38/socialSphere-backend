CREATE TABLE user_images(
                     id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                    user_id uuid NOT NULL,
                    image_id uuid NOT NULL,
                    created_at timestamp NOT NULL DEFAULT NOW()
);
