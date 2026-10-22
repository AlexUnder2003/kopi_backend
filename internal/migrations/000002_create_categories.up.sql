CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    icon VARCHAR(128) NOT NULL,
    user_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO categories (name, icon) VALUES
    ('Еда', 'hamburger'),
    ('Жильё', 'house'),
    ('ЖКХ', 'lightning'),
    ('Транспорт', 'car'),
    ('Одежда', 't-shirt'),
    ('Хобби', 'game-controller'),
    ('Развлечения', 'film-strip'),
    ('Здоровье', 'first-aid'),
    ('Образование', 'book-open'),
    ('Платежи', 'credit-card'),
    ('Кредиты', 'bank'),
    ('Подписки', 'device-mobile'),
    ('Связь', 'phone'),
    ('Путешествия', 'airplane-tilt'),
    ('Подарки', 'gift'),
    ('Покупки', 'shopping-cart'),
    ('Другое', 'package'),
    ('Переводы', 'transfer'),
    ('Зарплата', 'money'),
    ('Подработка', 'briefcase'),
    ('Инвестиции', 'chart-line-up'),
    ('Накопления', 'piggy-bank');

CREATE INDEX idx_categories_user_id ON categories(user_id);
CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION update_updated_at();