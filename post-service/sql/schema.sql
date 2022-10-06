CREATE TABLE post(
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    body   text NOT NULL,
    author      uuid NOT NULL,
    image_id uuid,
    created_at   timestamp NOT NULL
);