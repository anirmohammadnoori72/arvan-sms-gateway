CREATE TABLE IF NOT EXISTS wallets (
                                       user_id UUID PRIMARY KEY,
                                       balance BIGINT NOT NULL CHECK (balance >= 0),
                                       updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reservations (
                                            id UUID PRIMARY KEY,
                                            user_id UUID NOT NULL,
                                            amount BIGINT NOT NULL,
                                            used BOOLEAN DEFAULT FALSE,
                                            expires_at TIMESTAMP NOT NULL,
                                            created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reservations_expired ON reservations (used, expires_at);

CREATE TABLE IF NOT EXISTS messages (
                                        message_id UUID PRIMARY KEY,
                                        user_id UUID NOT NULL,
                                        phone_number TEXT NOT NULL,
                                        message TEXT NOT NULL,
                                        cost BIGINT NOT NULL,
                                        status TEXT NOT NULL CHECK (status IN ('queued', 'sent', 'failed', 'rejected')),
                                        created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE users (
                       id UUID PRIMARY KEY,
                       balance BIGINT NOT NULL DEFAULT 0,
                       is_vip BOOLEAN NOT NULL DEFAULT FALSE,
                       created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO users (id, balance, is_vip)
VALUES
    ('11111111-1111-1111-1111-111111111111', 10000000, TRUE),
    ('22222222-2222-2222-2222-222222222222', 10000000, FALSE);

CREATE INDEX IF NOT EXISTS idx_reservations_user ON reservations(user_id);
CREATE INDEX IF NOT EXISTS idx_messages_user ON messages(user_id);