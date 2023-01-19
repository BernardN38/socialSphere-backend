CREATE TABLE post(
    id         serial PRIMARY KEY DEFAULT gen_random_uuid(),
    body   text NOT NULL,
    user_id      int NOT NULL,
    author_name text NOT NULL,
    image_id uuid,
    created_at   timestamp NOT NULL DEFAULT NOW()
);
CREATE TABLE post_like(
                           post_id int,
                           user_id int NOT NULL,
                           PRIMARY KEY (post_id,user_id),
                           CONSTRAINT fk_post
                               FOREIGN KEY(post_id)
                                   REFERENCES post(id)
                                   ON DELETE CASCADE
);

CREATE TABLE post_comment(
    id         serial PRIMARY KEY,
    post_id      int NOT NULL REFERENCES post ON DELETE CASCADE, 
    comment_id int NOT NULL REFERENCES comment ON DELETE CASCADE
);

CREATE TABLE comment(
    id         serial PRIMARY KEY,
    body   text NOT NULL,
    user_id      int NOT NULL,
    author_name text NOT NULL,
    created_at   timestamp NOT NULL DEFAULT NOW()
);