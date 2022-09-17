package rss

var DBSchemas = []string{`
CREATE TABLE IF NOT EXISTS rss_feeds (
	id BIGSERIAL PRIMARY KEY,
	guild_id BIGINT NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,

	enabled BOOLEAN NOT NULL,
	channel_id BIGINT NOT NULL,
	mention_role BIGINT NOT NULL,

	feed_name TEXT NOT NULL,
	feed_title TEXT NOT NULL,
	feed_url TEXT NOT NULL
);
`, `
CREATE INDEX IF NOT EXISTS rss_feeds_guild_idx ON rss_feeds(guild_id);
`, `
CREATE INDEX IF NOT EXISTS rss_feeds_feeds_idx ON rss_feeds(feed_url);
`, `
CREATE TABLE IF NOT EXISTS rss_announcements (
	guild_id BIGINT PRIMARY KEY,
	announcement TEXT NOT NULL,
	enabled BOOLEAN NOT NULL DEFAULT false
);
`, `
CREATE INDEX IF NOT EXISTS rss_announcements_guild_idx ON rss_announcements(guild_id);
`,
}
