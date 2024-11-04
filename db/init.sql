CREATE TABLE IF NOT EXISTS clients (
                                       id SERIAL PRIMARY KEY,
                                       address VARCHAR(255) NOT NULL UNIQUE
    );