INSERT INTO users (id, name, is_seller)
SELECT 1, 'kauan', false
WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = 1);

INSERT INTO users (id, name, is_seller)
SELECT 2, 'adriana', true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = 2);

INSERT INTO users (id, name, is_seller)
SELECT 3, 'bruna', true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = 3);

INSERT INTO users (id, name, is_seller)
SELECT 4, 'cleiton', false
WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = 4);

INSERT INTO users (id, name, is_seller)
SELECT 5, 'lucas', false
WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = 5);

INSERT INTO users (id, name, is_seller)
SELECT 6, 'joao', true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = 6);

INSERT INTO users (id, name, is_seller)
SELECT 7, 'pedro', true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = 7);

-- Posts de exemplo (para demo no frontend)
-- Insere apenas se a tabela posts estiver vazia
INSERT INTO posts (
  user_id, date, date_str,
  product_id, product_name, type, brand, color, notes, image_url,
  category, price, has_promo, discount
)
SELECT
  6, NOW() - INTERVAL '2 days', to_char(NOW() - INTERVAL '2 days','DD-MM-YYYY'),
  101, 'Fone Bluetooth', 'audio', 'MeliSound', 'Preto', 'Excelente estado', '',
  1, 199.90, true, 15
WHERE NOT EXISTS (SELECT 1 FROM posts);

INSERT INTO posts (
  user_id, date, date_str,
  product_id, product_name, type, brand, color, notes, image_url,
  category, price, has_promo, discount
)
SELECT
  7, NOW() - INTERVAL '5 days', to_char(NOW() - INTERVAL '5 days','DD-MM-YYYY'),
  102, 'Mouse Gamer', 'periferico', 'ClickPro', 'Vermelho', 'RGB', '',
  2, 149.00, false, 0
WHERE (SELECT COUNT(*) FROM posts) = 1;