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
);`

	tableFollowers = `
CREATE TABLE IF NOT EXISTS followers
(
    id_speaker   SERIAL NOT NULL,--REFERENCES users(id)
    id_follower SERIAL NOT NULL,--REFERENCES users(id)
    UNIQUE (id_speaker, id_follower)
);`

	tableTags = `
CREATE TABLE IF NOT EXISTS tags
(
    id           SERIAL PRIMARY KEY,
    id_tag_maker SERIAL             NOT NULL, ----REFERENCES users(id)
    tag_name     VARCHAR(25) UNIQUE NOT NULL,
    created_at   TIMESTAMP          NOT NULL
);`

	tableArticles = `
CREATE TABLE IF NOT EXISTS articles
(
    id          SERIAL PRIMARY KEY,
    slug        VARCHAR(50) UNIQUE NOT NULL,
    title       VARCHAR(255)       NOT NULL,
    id_autor    SERIAL             NOT NULL, --REFERENCES users(id)
    description VARCHAR(2048)      NOT NULL,
    body        VARCHAR(2048)      NOT NULL,
    created_at  TIMESTAMP          NOT NULL,
    updated_at  TIMESTAMP          NULL
);`

	tableArticleFavorite = `
CREATE TABLE IF NOT EXISTS article_favorite
(
    id_user    SERIAL NOT NULL, --REFERENCES users (id)
    id_article SERIAL NOT NULL, --REFERENCES articles (id)
    UNIQUE (id_user, id_article)
);`

	tableArticleTags = `
CREATE TABLE IF NOT EXISTS articles_tags
(
    id_article SERIAL NOT NULL, --REFERENCES articles (id)
    id_tag     SERIAL NOT NULL, --REFERENCES tags (id)
    UNIQUE (id_article, id_tag)
);`

	tableComments = `
CREATE TABLE IF NOT EXISTS comments
(
    id         BIGSERIAL PRIMARY KEY,
    id_autor   SERIAL        NOT NULL, --REFERENCES users (id)
    id_article SERIAL        NOT NULL, --REFERENCES articles (id)
    body       VARCHAR(2048) NOT NULL,
    created_at TIMESTAMP     NOT NULL,
    updated_at TIMESTAMP     NULL
);`
)
