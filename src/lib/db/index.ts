import Database from "better-sqlite3";
import { drizzle } from "drizzle-orm/better-sqlite3";
import path from "node:path";
import * as schema from "./schema";

declare global {
	// eslint-disable-next-line no-var
	var __db: ReturnType<typeof drizzle> | undefined;
}

function createDb() {
	const dbPath = path.join(process.cwd(), "drizzle", "reconta.db");
	const sqlite = new Database(dbPath);
	sqlite.pragma("journal_mode = WAL");
	sqlite.pragma("foreign_keys = ON");
	return drizzle(sqlite, { schema });
}

export const db = globalThis.__db ?? createDb();

if (process.env.NODE_ENV !== "production") {
	globalThis.__db = db;
}

export type DB = typeof db;
