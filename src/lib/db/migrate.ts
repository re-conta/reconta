import Database from "better-sqlite3";
import { drizzle } from "drizzle-orm/better-sqlite3";
import { migrate } from "drizzle-orm/better-sqlite3/migrator";
import path from "node:path";

const dbPath = path.join(process.cwd(), "drizzle", "reconta.db");
const sqlite = new Database(dbPath);
const db = drizzle(sqlite);

migrate(db, { migrationsFolder: path.join(process.cwd(), "drizzle") });
console.log("Migration completed");
sqlite.close();
