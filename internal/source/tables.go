package source

const (
	tableUsers = `
CREATE TABLE IF NOT EXISTS users
(
    id              SERIAL PRIMARY KEY,
    login           VARCHAR(128) UNIQUE NOT NULL,
    hash_password   VARCHAR(80)         NOT NULL,
    access          VARCHAR(1)          NOT NULL,
    first_name      VARCHAR(128)        NOT NULL,
    last_name       VARCHAR(128)        NULL,
    phone           VARCHAR(20) UNIQUE  NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    image           VARCHAR(512)        NULL,
    biography       VARCHAR(2048)       NULL,
    created_at      TIMESTAMP           NOT NULL,
    updated_at      TIMESTAMP           NULL,
    last_connection TIMESTAMP           NULL
)`
	tableFollowers = `
CREATE TABLE IF NOT EXISTS followers
(
    id_speaker   SERIAL NOT NULL,--REFERENCES users(id)
    id_follower SERIAL NOT NULL,--REFERENCES users(id)
    UNIQUE (id_speaker, id_follower)
);`
)
