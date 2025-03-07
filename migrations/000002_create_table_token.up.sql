CREATE TABLE
  tokens (
    id BIGSERIAL PRIMARY KEY,
    value TEXT NOT NULL,
    token_id UUID NOT NULL,
    is_revoked BOOLEAN DEFAULT false,
    user_id UUID NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMP(0)
    WITH
      TIME ZONE NOT NULL DEFAULT NOW (),
      expired_at TIMESTAMP(0)
    WITH
      TIME ZONE NOT NULL DEFAULT NOW ()
  );

CREATE INDEX idx_token_user ON tokens (user_id);

CREATE INDEX idx_token_id ON tokens (token_id);