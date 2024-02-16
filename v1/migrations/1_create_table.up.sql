CREATE TABLE banners (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    age_start INTEGER,
    age_end INTEGER,
    gender CHAR(1)[],
    country VARCHAR(10)[],
    platform VARCHAR(127)[]
);