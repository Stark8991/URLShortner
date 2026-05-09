-- URL Shortener Database Schema for PostgreSQL

-- Create the urls table
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    click_count INTEGER DEFAULT 0
);


-- Create an index on created_at for cleanup queries
CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at);

-- Optional: Create a partial index for non-expired URLs
CREATE INDEX IF NOT EXISTS idx_urls_active ON urls(short_code)
WHERE expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP;