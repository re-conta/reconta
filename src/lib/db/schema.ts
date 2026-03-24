import { sql } from "drizzle-orm";
import { integer, real, sqliteTable, text } from "drizzle-orm/sqlite-core";

// ─── Better Auth tables ───────────────────────────────────────────────────────

export const user = sqliteTable("user", {
	id: text("id").primaryKey(),
	name: text("name").notNull(),
	email: text("email").notNull().unique(),
	emailVerified: integer("email_verified", { mode: "boolean" })
		.notNull()
		.default(false),
	image: text("image"),
	createdAt: integer("created_at", { mode: "timestamp" }).notNull(),
	updatedAt: integer("updated_at", { mode: "timestamp" }).notNull(),
});

export const session = sqliteTable("session", {
	id: text("id").primaryKey(),
	expiresAt: integer("expires_at", { mode: "timestamp" }).notNull(),
	token: text("token").notNull().unique(),
	createdAt: integer("created_at", { mode: "timestamp" }).notNull(),
	updatedAt: integer("updated_at", { mode: "timestamp" }).notNull(),
	ipAddress: text("ip_address"),
	userAgent: text("user_agent"),
	userId: text("user_id")
		.notNull()
		.references(() => user.id, { onDelete: "cascade" }),
});

export const authAccount = sqliteTable("account", {
	id: text("id").primaryKey(),
	accountId: text("account_id").notNull(),
	providerId: text("provider_id").notNull(),
	userId: text("user_id")
		.notNull()
		.references(() => user.id, { onDelete: "cascade" }),
	accessToken: text("access_token"),
	refreshToken: text("refresh_token"),
	idToken: text("id_token"),
	accessTokenExpiresAt: integer("access_token_expires_at", {
		mode: "timestamp",
	}),
	refreshTokenExpiresAt: integer("refresh_token_expires_at", {
		mode: "timestamp",
	}),
	scope: text("scope"),
	password: text("password"),
	createdAt: integer("created_at", { mode: "timestamp" }).notNull(),
	updatedAt: integer("updated_at", { mode: "timestamp" }).notNull(),
});

export const verification = sqliteTable("verification", {
	id: text("id").primaryKey(),
	identifier: text("identifier").notNull(),
	value: text("value").notNull(),
	expiresAt: integer("expires_at", { mode: "timestamp" }).notNull(),
	createdAt: integer("created_at", { mode: "timestamp" }),
	updatedAt: integer("updated_at", { mode: "timestamp" }),
});

// ─── App tables ───────────────────────────────────────────────────────────────

export const accounts = sqliteTable("accounts", {
	id: integer("id").primaryKey({ autoIncrement: true }),
	userId: text("user_id")
		.notNull()
		.references(() => user.id, { onDelete: "cascade" }),
	name: text("name").notNull(),
	type: text("type", { enum: ["checking", "savings", "credit", "investment"] })
		.notNull()
		.default("checking"),
	balance: real("balance").notNull().default(0),
	createdAt: text("created_at").notNull().default(sql`(datetime('now'))`),
});

export const categories = sqliteTable("categories", {
	id: integer("id").primaryKey({ autoIncrement: true }),
	userId: text("user_id")
		.notNull()
		.references(() => user.id, { onDelete: "cascade" }),
	name: text("name").notNull(),
	color: text("color").notNull().default("#6366f1"),
	icon: text("icon").notNull().default("circle"),
	type: text("type", { enum: ["income", "expense", "both"] })
		.notNull()
		.default("both"),
});

export const transactions = sqliteTable("transactions", {
	id: integer("id").primaryKey({ autoIncrement: true }),
	userId: text("user_id")
		.notNull()
		.references(() => user.id, { onDelete: "cascade" }),
	date: text("date").notNull(),
	description: text("description").notNull(),
	amount: real("amount").notNull(),
	type: text("type", { enum: ["income", "expense"] }).notNull(),
	categoryId: integer("category_id").references(() => categories.id),
	accountId: integer("account_id").references(() => accounts.id),
	notes: text("notes"),
	importedFrom: text("imported_from"),
	createdAt: text("created_at").notNull().default(sql`(datetime('now'))`),
});

export const bills = sqliteTable("bills", {
	id: integer("id").primaryKey({ autoIncrement: true }),
	userId: text("user_id")
		.notNull()
		.references(() => user.id, { onDelete: "cascade" }),
	name: text("name").notNull(),
	amount: real("amount").notNull(),
	dueDay: integer("due_day").notNull(),
	categoryId: integer("category_id").references(() => categories.id),
	isActive: integer("is_active", { mode: "boolean" }).notNull().default(true),
	createdAt: text("created_at").notNull().default(sql`(datetime('now'))`),
});

export const billPayments = sqliteTable("bill_payments", {
	id: integer("id").primaryKey({ autoIncrement: true }),
	billId: integer("bill_id")
		.notNull()
		.references(() => bills.id, { onDelete: "cascade" }),
	month: integer("month").notNull(),
	year: integer("year").notNull(),
	isPaid: integer("is_paid", { mode: "boolean" }).notNull().default(false),
	paidAt: text("paid_at"),
	amount: real("amount"),
});

export const pdfImports = sqliteTable("pdf_imports", {
	id: integer("id").primaryKey({ autoIncrement: true }),
	userId: text("user_id")
		.notNull()
		.references(() => user.id, { onDelete: "cascade" }),
	filename: text("filename").notNull(),
	accountId: integer("account_id").references(() => accounts.id),
	transactionCount: integer("transaction_count").notNull().default(0),
	importedAt: text("imported_at").notNull().default(sql`(datetime('now'))`),
});

// ─── Types ────────────────────────────────────────────────────────────────────

export type User = typeof user.$inferSelect;
export type Account = typeof accounts.$inferSelect;
export type NewAccount = typeof accounts.$inferInsert;
export type Category = typeof categories.$inferSelect;
export type NewCategory = typeof categories.$inferInsert;
export type Transaction = typeof transactions.$inferSelect;
export type NewTransaction = typeof transactions.$inferInsert;
export type Bill = typeof bills.$inferSelect;
export type NewBill = typeof bills.$inferInsert;
export type BillPayment = typeof billPayments.$inferSelect;
export type NewBillPayment = typeof billPayments.$inferInsert;
