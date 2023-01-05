CREATE TABLE post(
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    body   text NOT NULL,
    user_id      uuid NOT NULL,
    author_name text NOT NULL,
    image_id uuid,
    created_at   timestamp NOT NULL DEFAULT NOW()
);
CREATE TABLE post_like(
                           post_id uuid,
                           user_id uuid NOT NULL,
                           PRIMARY KEY (post_id,user_id),
                           CONSTRAINT fk_post
                               FOREIGN KEY(post_id)
                                   REFERENCES post(id)
                                   ON DELETE CASCADE
);

CREATE TABLE post_comment(
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id      uuid NOT NULL REFERENCES post ON DELETE CASCADE, 
    comment_id uuid NOT NULL REFERENCES comment ON DELETE CASCADE
);

CREATE TABLE comment(
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    body   text NOT NULL,
    user_id      uuid NOT NULL,
    author_name text NOT NULL,
    created_at   timestamp NOT NULL DEFAULT NOW()
);