CREATE TABLE `hotspots` (
	`id` text PRIMARY KEY NOT NULL,
	`title` text NOT NULL,
	`content` text NOT NULL,
	`url` text NOT NULL,
	`source` text NOT NULL,
	`source_id` text,
	`is_real` integer DEFAULT true NOT NULL,
	`relevance` integer DEFAULT 0 NOT NULL,
	`relevance_reason` text,
	`keyword_mentioned` integer,
	`importance` text DEFAULT 'low' NOT NULL,
	`summary` text,
	`view_count` integer,
	`like_count` integer,
	`retweet_count` integer,
	`reply_count` integer,
	`comment_count` integer,
	`quote_count` integer,
	`danmaku_count` integer,
	`author_name` text,
	`author_username` text,
	`author_avatar` text,
	`author_followers` integer,
	`author_verified` integer,
	`published_at` integer,
	`created_at` integer DEFAULT CURRENT_TIMESTAMP NOT NULL,
	`keyword_id` text,
	FOREIGN KEY (`keyword_id`) REFERENCES `keywords`(`id`) ON UPDATE no action ON DELETE set null
);
--> statement-breakpoint
CREATE UNIQUE INDEX `url_source_idx` ON `hotspots` (`url`,`source`);--> statement-breakpoint
CREATE TABLE `keywords` (
	`id` text PRIMARY KEY NOT NULL,
	`text` text NOT NULL,
	`category` text,
	`is_active` integer DEFAULT true NOT NULL,
	`created_at` integer DEFAULT CURRENT_TIMESTAMP NOT NULL,
	`updated_at` integer DEFAULT CURRENT_TIMESTAMP NOT NULL
);
--> statement-breakpoint
CREATE UNIQUE INDEX `keywords_text_unique` ON `keywords` (`text`);--> statement-breakpoint
CREATE TABLE `notifications` (
	`id` text PRIMARY KEY NOT NULL,
	`type` text NOT NULL,
	`title` text NOT NULL,
	`content` text NOT NULL,
	`is_read` integer DEFAULT false NOT NULL,
	`hotspot_id` text,
	`created_at` integer DEFAULT CURRENT_TIMESTAMP NOT NULL
);
--> statement-breakpoint
CREATE TABLE `settings` (
	`id` text PRIMARY KEY NOT NULL,
	`key` text NOT NULL,
	`value` text NOT NULL
);
--> statement-breakpoint
CREATE UNIQUE INDEX `settings_key_unique` ON `settings` (`key`);