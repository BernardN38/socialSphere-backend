#!/bin/bash
set -e
export PGPASSWORD=$POSTGRES_PASSWORD;
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  CREATE USER $APP_DB_USER WITH PASSWORD '$APP_DB_PASS';
  CREATE DATABASE $AUTH_DB_NAME;
  CREATE DATABASE $POST_DB_NAME;
  CREATE DATABASE $IMAGE_DB_NAME;
  GRANT ALL PRIVILEGES ON DATABASE $AUTH_DB_NAME TO $APP_DB_USER;
  GRANT ALL PRIVILEGES ON DATABASE $POST_DB_NAME TO $APP_DB_USER;
  \connect $AUTH_DB_NAME $APP_DB_USER
  BEGIN;
    CREATE TABLE users
    (
        id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
        username   text NOT NULL UNIQUE,
        email      text NOT NULL UNIQUE,
        password   text NOT NULL,
        first_name text NOT NULL,
        last_name  text NOT NULL
    );
  COMMIT;
  \connect $POST_DB_NAME $APP_DB_USER
    BEGIN;
     CREATE TABLE post(
         id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
         body   text NOT NULL,
         author      uuid NOT NULL,
           author_name text NOT NULL,
         image_id uuid,
         created_at   timestamp NOT NULL
     );
    COMMIT;
  \connect $IMAGE_DB_NAME $APP_DB_USER
    BEGIN;
    CREATE TABLE user_images(
                         id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                        user_id uuid NOT NULL,
                        image_id uuid NOT NULL,
                        created_at timestamp NOT NULL DEFAULT NOW()
    );
    COMMIT;
EOSQL
