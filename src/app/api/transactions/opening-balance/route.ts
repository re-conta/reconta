export const dynamic = "force-dynamic";

import { and, eq, sql } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { accounts, monthlyOpeningBalances } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

async function getAccountsBalance(userId: string) {
	const [row] = await db
		.select({ total: sql<number>`coalesce(sum(${accounts.balance}), 0)` })
		.from(accounts)
		.where(eq(accounts.userId, userId));
	return Number(row?.total ?? 0);
}

export async function GET(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { searchParams } = new URL(request.url);
	const month = Number(searchParams.get("month"));
	const year = Number(searchParams.get("year"));

	if (!month || !year) {
		return NextResponse.json(
			{ error: "month and year are required" },
			{ status: 400 },
		);
	}

	const [row] = await db
		.select()
		.from(monthlyOpeningBalances)
		.where(
			and(
				eq(monthlyOpeningBalances.userId, userId),
				eq(monthlyOpeningBalances.month, month),
				eq(monthlyOpeningBalances.year, year),
			),
		)
		.limit(1);

	if (row) return NextResponse.json({ amount: row.amount });

	const amount = await getAccountsBalance(userId);
	return NextResponse.json({ amount });
}

export async function POST(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const body = await request.json();
	const month = Number(body.month);
	const year = Number(body.year);
	const amount = Number(body.amount);

	if (!month || !year || Number.isNaN(amount)) {
		return NextResponse.json(
			{ error: "month, year and amount are required" },
			{ status: 400 },
		);
	}

	const [existing] = await db
		.select({ id: monthlyOpeningBalances.id })
		.from(monthlyOpeningBalances)
		.where(
			and(
				eq(monthlyOpeningBalances.userId, userId),
				eq(monthlyOpeningBalances.month, month),
				eq(monthlyOpeningBalances.year, year),
			),
		)
		.limit(1);

	if (existing) {
		await db
			.update(monthlyOpeningBalances)
			.set({
				amount,
				updatedAt: new Date().toISOString().replace("T", " ").slice(0, 19),
			})
			.where(eq(monthlyOpeningBalances.id, existing.id));
	} else {
		await db.insert(monthlyOpeningBalances).values({
			userId,
			month,
			year,
			amount,
		});
	}

	return NextResponse.json({ amount });
}
