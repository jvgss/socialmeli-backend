-- Adiciona imagem do produto nos posts
ALTER TABLE posts ADD COLUMN IF NOT EXISTS image_url TEXT;

-- Default para registros existentes
UPDATE posts SET image_url = '' WHERE image_url IS NULL;
