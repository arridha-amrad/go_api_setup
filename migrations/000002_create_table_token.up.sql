CREATE TABLE
  tokens (
    id BIGSERIAL PRIMARY KEY,
    hash TEXT NOT NULL,
    is_revoked BOOLEAN DEFAULT false,
    device_id UUID NOT NULL,
    user_id UUID NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    expired_at TIMESTAMP(0)
    WITH
      TIME ZONE NOT NULL DEFAULT NOW () + INTERVAL '365 days'
  );

CREATE INDEX idx_token_user_device ON tokens (user_id, device_id);