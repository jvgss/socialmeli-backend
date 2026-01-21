CREATE TABLE IF NOT EXISTS users (
  id        INT PRIMARY KEY,
  name      TEXT NOT NULL,
  is_seller BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS follows (
  user_id   INT NOT NULL,
  seller_id INT NOT NULL,
  PRIMARY KEY (user_id, seller_id),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (seller_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS posts (
  id           SERIAL PRIMARY KEY,
  user_id      INT NOT NULL,
  date         TIMESTAMP NOT NULL,
  date_str     TEXT NOT NULL,

  product_id   INT NOT NULL,
  product_name TEXT NOT NULL,
  type         TEXT NOT NULL,
  brand        TEXT NOT NULL,
  color        TEXT NOT NULL,
  notes        TEXT NOT NULL,

  category     INT NOT NULL,
  price        NUMERIC(12,2) NOT NULL,
  has_promo    BOOLEAN NOT NULL DEFAULT FALSE,
  discount     NUMERIC(12,2) NOT NULL DEFAULT 0,

  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_posts_user_date ON posts(user_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_posts_promo ON posts(has_promo);
