CREATE TABLE IF NOT EXISTS items (
                                      id bigserial PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL,
    price integer NOT NULL,
    category text[] NOT NULL
    );