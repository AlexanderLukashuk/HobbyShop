CREATE TABLE IF NOT EXISTS baskets (
                                     id bigserial PRIMARY KEY,
                                     items INTEGER[]  NOT NULL,
                                     user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE
);