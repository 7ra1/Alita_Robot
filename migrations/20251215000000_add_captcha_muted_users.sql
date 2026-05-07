-- Add captcha_muted_users table for tracking users to auto-unmute
CREATE TABLE IF NOT EXISTS captcha_muted_users (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    chat_id BIGINT NOT NULL,
    unmute_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_captcha_muted_user_chat ON captcha_muted_users(user_id, chat_id);
CREATE INDEX IF NOT EXISTS idx_captcha_unmute_at ON captcha_muted_users(unmute_at);

-- Add foreign key constraints.
-- NOTE: The migration runner wraps ALTER TABLE ... ADD CONSTRAINT in an
-- idempotent DO block during SQL cleaning, so keep these as plain statements.
ALTER TABLE captcha_muted_users
ADD CONSTRAINT fk_captcha_muted_user
FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;

ALTER TABLE captcha_muted_users
ADD CONSTRAINT fk_captcha_muted_chat
FOREIGN KEY (chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE;
