package yageconomy

var DBSchemas = []string{`
CREATE TABLE IF NOT EXISTS economy_configs (
	guild_id BIGINT PRIMARY KEY,
	enabled BOOLEAN NOT NULL,

	admins BIGINT[],
	
	currency_name TEXT NOT NULL,
	currency_name_plural TEXT NOT NULL,
	currency_symbol TEXT NOT NULL,

	daily_frequency BIGINT NOT NULL,
	daily_amount BIGINT NOT NULL,

	chatmoney_frequency BIGINT NOT NULL,
	chatmoney_amount_min BIGINT NOT NULL,
	chatmoney_amount_max BIGINT NOT NULL,

	auto_plant_channels BIGINT[],
	auto_plant_min BIGINT NOT NULL,
	auto_plant_max BIGINT NOT NULL,
	auto_plant_chance NUMERIC(5,4) NOT NULL,

	start_balance BIGINT NOT NULL,

	fishing_max_win_amount BIGINT NOT NULL,
	fishing_min_win_amount BIGINT NOT NULL,
	fishing_cooldown INT NOT NULL,

	rob_fine INT NOT NULL,
	rob_cooldown INT NOT NULL
);
`, `
ALTER TABLE economy_configs ADD COLUMN IF NOT EXISTS heist_server_cooldown INT NOT NULL DEFAULT 0;
`, `
ALTER TABLE economy_configs ADD COLUMN IF NOT EXISTS heist_failed_gambling_ban_duration INT NOT NULL DEFAULT 0;
`, `
ALTER TABLE economy_configs ADD COLUMN IF NOT EXISTS heist_last_usage TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() - interval '10 days';
`, `
ALTER TABLE economy_configs ADD COLUMN IF NOT EXISTS heist_fixed_payout INT NOT NULL DEFAULT 0;
`, `
ALTER TABLE economy_configs ADD COLUMN IF NOT EXISTS enabled_channels BIGINT[];
`, `

CREATE TABLE IF NOT EXISTS economy_users (
	guild_id BIGINT NOT NULL,
	user_id BIGINT NOT NULL,

	money_bank BIGINT NOT NULL,
	money_wallet BIGINT NOT NULL,

	last_daily_claim TIMESTAMP WITH TIME ZONE NOT NULL,
	last_chatmoney_claim TIMESTAMP WITH TIME ZONE NOT NULL,
 	last_fishing TIMESTAMP WITH TIME ZONE NOT NULL,

 	-- who's claimed this user as a waifu
	waifud_by BIGINT NOT NULL,

	-- the waifus this user has claimed
	waifus BIGINT[],

	-- the items and item worth this user has been gifted
	waifu_item_worth BIGINT NOT NULL,
	waifu_last_claim_amount BIGINT NOT NULL,
	waifu_extra_worth BIGINT NOT NULL,

	waifu_affinity_towards BIGINT NOT NULL,
	waifu_divorces INT NOT NULL,
	waifu_affinity_changes INT NOT NULL,

	fish_caugth BIGINT NOT NULL,

	gambling_boost_percentage INT NOT NULL,

	last_interest_update TIMESTAMP WITH TIME ZONE NOT NULL,
	last_rob_attempt TIMESTAMP WITH TIME ZONE NOT NULL,

	PRIMARY KEY(guild_id, user_id)
);

`, `
ALTER TABLE economy_users ADD COLUMN IF NOT EXISTS last_failed_heist TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() - interval '10 days';
`, `

CREATE TABLE IF NOT EXISTS economy_waifu_items (
	guild_id BIGINT NOT NULL,
	local_id BIGINT NOT NULL,

	name TEXT NOT NULL,
	icon TEXT NOT NULL,
	price INT NOT NULL,

	PRIMARY KEY(guild_id, local_id)
);

ALTER TABLE economy_waifu_items ADD COLUMN IF NOT EXISTS gambling_boost INT NOT NULL DEFAULT 0;
`, `

CREATE TABLE IF NOT EXISTS economy_logs (
	id BIGSERIAL PRIMARY KEY,
	guild_id BIGINT NOT NULL,

	author_id BIGINT NOT NULL,
	target_id BIGINT NOT NULL,

	amount BIGINT NOT NULL,

	action SMALLINT NOT NULL
);
`, `

CREATE INDEX IF NOT EXISTS economy_logs_guild_target_idx ON economy_logs (guild_id, target_id);
`, `

CREATE TABLE IF NOT EXISTS economy_shop_items (
	guild_id BIGINT NOT NULL,
	local_id BIGINT NOT NULL,

	icon TEXT NOT NULL,
	name TEXT NOT NULL,
	role_id BIGINT NOT NULL,
	gambling_boost_percentage INT NOT NULL,

	cost BIGINT NOT NULL,
	type SMALLINT NOT NULL, -- 0 for roles, 1 for lists

	PRIMARY KEY (guild_id, local_id)
);

`, `
CREATE TABLE IF NOT EXISTS economy_shop_list_items (
	guild_id BIGINT NOT NULL,
	local_id BIGINT NOT NULL,

	list_id BIGINT NOT NULL,

	value TEXT NOT NULL,
	purchased_by BIGINT NOT NULL,

	PRIMARY KEY (guild_id, local_id)
);

`, `
CREATE TABLE IF NOT EXISTS economy_plants (
	message_id BIGINT PRIMARY KEY,
	channel_id BIGINT NOT NULL,
	guild_id BIGINT NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE NOT NULL,

	author_id BIGINT NOT NULL,
	amount BIGINT NOT NULL,

	password TEXT NOT NULL
);

`, `
CREATE TABLE IF NOT EXISTS economy_pick_images2 (
	id BIGSERIAL PRIMARY KEY,
	guild_id BIGINT NOT NULL,
	image BYTEA NOT NULL
);
`, `

CREATE TABLE IF NOT EXISTS economy_users_waifu_items (
	guild_id BIGINT NOT NULL,
	user_id BIGINT NOT NULL,
	item_id BIGINT NOT NULL,

	quantity INT NOT NULL,

	FOREIGN KEY (guild_id, user_id) REFERENCES economy_users (guild_id, user_id) ON DELETE CASCADE,
	FOREIGN KEY (guild_id, item_id) REFERENCES economy_waifu_items (guild_id, local_id) ON DELETE CASCADE,

	PRIMARY KEY(guild_id, user_id, item_id)
);

`}
