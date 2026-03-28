/**
 * Bill notification cron job.
 * Run via: pnpm tsx src/db/cron.ts
 * Scheduled by the reconta-cron.timer systemd unit.
 */

import "dotenv/config";
import path from "node:path";
import { and, eq, sql } from "drizzle-orm";
import Database from "better-sqlite3";
import { drizzle } from "drizzle-orm/better-sqlite3";
import * as schema from "../lib/db/schema";
import { sendBillNotificationEmail } from "../lib/email";
import { sendTextMessage } from "../lib/whatsapp";

const { user, bills, billPayments, notificationSettings, notificationLogs } =
	schema;

function createDb() {
	const dbPath = path.join(process.cwd(), "reconta.db");
	const sqlite = new Database(dbPath);
	sqlite.pragma("journal_mode = WAL");
	sqlite.pragma("foreign_keys = ON");
	return drizzle(sqlite, { schema });
}

function formatCurrency(amount: number): string {
	return new Intl.NumberFormat("pt-BR", {
		style: "currency",
		currency: "BRL",
	}).format(amount);
}

async function main() {
	const db = createDb();
	const today = new Date();
	const currentMonth = today.getMonth() + 1;
	const currentYear = today.getFullYear();
	const todayDay = today.getDate();
	const daysInMonth = new Date(currentYear, currentMonth, 0).getDate();

	const appUrl = process.env.APP_URL ?? "https://reconta.app";
	const settingsUrl = `${appUrl}/ajustes`;

	console.log(
		`[cron] Checking bills for ${currentYear}-${String(currentMonth).padStart(2, "0")}-${String(todayDay).padStart(2, "0")}`,
	);

	// Load all users with notifications enabled
	const settings = await db
		.select({
			userId: notificationSettings.userId,
			emailAddress: notificationSettings.emailAddress,
			enabled: notificationSettings.enabled,
			whatsappEnabled: notificationSettings.whatsappEnabled,
			whatsappNumber: notificationSettings.whatsappNumber,
			daysBeforeDue: notificationSettings.daysBeforeDue,
			daysAfterDue: notificationSettings.daysAfterDue,
			maxNotificationsPerBill: notificationSettings.maxNotificationsPerBill,
			intervalDays: notificationSettings.intervalDays,
			userName: user.name,
			userEmail: user.email,
		})
		.from(notificationSettings)
		.innerJoin(user, eq(notificationSettings.userId, user.id))
		.where(eq(notificationSettings.enabled, true));

	console.log(
		`[cron] Found ${settings.length} user(s) with notifications enabled`,
	);

	for (const setting of settings) {
		const emailTo = setting.emailAddress ?? setting.userEmail;

		// Load active bills for this user
		const userBills = await db
			.select({
				id: bills.id,
				name: bills.name,
				amount: bills.amount,
				dueDay: bills.dueDay,
				paymentId: billPayments.id,
				isPaid: billPayments.isPaid,
			})
			.from(bills)
			.leftJoin(
				billPayments,
				and(
					eq(billPayments.billId, bills.id),
					eq(billPayments.month, currentMonth),
					eq(billPayments.year, currentYear),
				),
			)
			.where(and(eq(bills.userId, setting.userId), eq(bills.isActive, true)));

		const overdueBills: Array<{
			id: number;
			name: string;
			dueDay: number;
			daysOverdue: number;
			amountFormatted: string;
		}> = [];
		const upcomingBills: Array<{
			id: number;
			name: string;
			dueDay: number;
			daysUntil: number;
			amountFormatted: string;
		}> = [];

		for (const bill of userBills) {
			if (bill.isPaid) continue;

			// Clamp dueDay to actual days in month
			const effectiveDueDay = Math.min(bill.dueDay, daysInMonth);
			const daysUntil = effectiveDueDay - todayDay;

			if (daysUntil < 0 && Math.abs(daysUntil) <= setting.daysAfterDue) {
				// Overdue
				overdueBills.push({
					id: bill.id,
					name: bill.name,
					dueDay: effectiveDueDay,
					daysOverdue: Math.abs(daysUntil),
					amountFormatted: formatCurrency(bill.amount),
				});
			} else if (daysUntil >= 0 && daysUntil <= setting.daysBeforeDue) {
				// Upcoming
				upcomingBills.push({
					id: bill.id,
					name: bill.name,
					dueDay: effectiveDueDay,
					daysUntil,
					amountFormatted: formatCurrency(bill.amount),
				});
			}
		}

		if (overdueBills.length === 0 && upcomingBills.length === 0) {
			continue;
		}

		// Check notification caps for each relevant bill
		const relevantBills = [
			...overdueBills.map((b) => b.id),
			...upcomingBills.map((b) => b.id),
		];

		const shouldNotify = await Promise.all(
			relevantBills.map(async (billId) => {
				const logs = await db
					.select()
					.from(notificationLogs)
					.where(
						and(
							eq(notificationLogs.userId, setting.userId),
							eq(notificationLogs.billId, billId),
							eq(notificationLogs.month, currentMonth),
							eq(notificationLogs.year, currentYear),
						),
					);

				if (logs.length === 0) return { billId, send: true };

				const latest = logs[logs.length - 1];
				const totalSent = latest.notificationCount;

				if (totalSent >= setting.maxNotificationsPerBill) {
					return { billId, send: false };
				}

				// Check interval
				const lastSent = new Date(latest.sentAt);
				const hoursSince =
					(today.getTime() - lastSent.getTime()) / (1000 * 60 * 60);
				const required = setting.intervalDays * 24;
				return { billId, send: hoursSince >= required };
			}),
		);

		const billsToNotify = new Set(
			shouldNotify.filter((r) => r.send).map((r) => r.billId),
		);

		if (billsToNotify.size === 0) {
			console.log(`[cron] User ${setting.userId}: skipped (rate-limited)`);
			continue;
		}

		const filteredOverdue = overdueBills.filter((b) => billsToNotify.has(b.id));
		const filteredUpcoming = upcomingBills.filter((b) =>
			billsToNotify.has(b.id),
		);

		if (filteredOverdue.length === 0 && filteredUpcoming.length === 0) continue;

		try {
			await sendBillNotificationEmail(emailTo, {
				name: setting.userName,
				overdueBills: filteredOverdue,
				upcomingBills: filteredUpcoming,
				appUrl,
				settingsUrl,
			});

			// Send WhatsApp notification if enabled
			if (setting.whatsappEnabled && setting.whatsappNumber) {
				try {
					const lines: string[] = [];
					lines.push(`⚠️ *Alerta de contas — ReConta*`);
					lines.push("");

					if (filteredOverdue.length > 0) {
						lines.push("🔴 *Contas vencidas:*");
						for (const b of filteredOverdue) {
							lines.push(
								`• ${b.name} — dia ${b.dueDay} (${b.daysOverdue} dia${b.daysOverdue === 1 ? "" : "s"} atrás) — ${b.amountFormatted}`,
							);
						}
						lines.push("");
					}

					if (filteredUpcoming.length > 0) {
						lines.push("🟡 *Contas a vencer:*");
						for (const b of filteredUpcoming) {
							lines.push(
								`• ${b.name} — dia ${b.dueDay} (${b.daysUntil === 0 ? "hoje" : `em ${b.daysUntil} dia${b.daysUntil === 1 ? "" : "s"}`}) — ${b.amountFormatted}`,
							);
						}
						lines.push("");
					}

					lines.push(`Acesse: ${appUrl}/contas`);

					await sendTextMessage(setting.whatsappNumber, lines.join("\n"));
					console.log(`[cron] Sent WhatsApp to ${setting.whatsappNumber}`);
				} catch (whatsappErr) {
					console.error(
						`[cron] Failed to send WhatsApp to ${setting.whatsappNumber}:`,
						whatsappErr,
					);
				}
			}

			// Upsert notification log for each bill
			for (const billId of billsToNotify) {
				const existingLogs = await db
					.select()
					.from(notificationLogs)
					.where(
						and(
							eq(notificationLogs.userId, setting.userId),
							eq(notificationLogs.billId, billId),
							eq(notificationLogs.month, currentMonth),
							eq(notificationLogs.year, currentYear),
						),
					);

				if (existingLogs.length > 0) {
					await db
						.update(notificationLogs)
						.set({
							notificationCount:
								existingLogs[existingLogs.length - 1].notificationCount + 1,
							sentAt: sql`(datetime('now'))`,
						})
						.where(
							eq(notificationLogs.id, existingLogs[existingLogs.length - 1].id),
						);
				} else {
					await db.insert(notificationLogs).values({
						userId: setting.userId,
						billId,
						month: currentMonth,
						year: currentYear,
						notificationCount: 1,
					});
				}
			}

			console.log(
				`[cron] Sent notification to ${emailTo} (${filteredOverdue.length} overdue, ${filteredUpcoming.length} upcoming)`,
			);
		} catch (err) {
			console.error(`[cron] Failed to send email to ${emailTo}:`, err);
		}
	}

	console.log("[cron] Done");
	process.exit(0);
}

main().catch((err) => {
	console.error("[cron] Fatal error:", err);
	process.exit(1);
});
