ALTER TABLE `notification_settings` ADD `whatsapp_enabled` integer DEFAULT false NOT NULL;--> statement-breakpoint
ALTER TABLE `notification_settings` ADD `whatsapp_number` text;
