export const dynamic = "force-dynamic";

import { and, desc, eq, gte, lte, sql } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { categories, transactions } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";
import { checkSharedAccess } from "@/lib/shared-access";

export async function GET(
	request: Request,
	{ params }: { params: Promise<{ ownerId: string }> },
) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { ownerId } = await params;

	const access = await checkSharedAccess(ownerId, userId);
	if (!access) {
		return NextResponse.json({ error: "Acesso negado" }, { status: 403 });
	}

	const { searchParams } = new URL(request.url);
	const scope = searchParams.get("scope") ?? "all";
	const month = searchParams.get("month");
	const year = searchParams.get("year");

	const conditions = [eq(transactions.userId, ownerId)];

	if (scope === "month" && month && year) {
		const start = new Date(Number(year), Number(month) - 1, 1)
			.toISOString()
			.split("T")[0];
		const end = new Date(Number(year), Number(month), 0)
			.toISOString()
			.split("T")[0];
		conditions.push(gte(transactions.date, start), lte(transactions.date, end));
	} else if (scope === "year" && year) {
		conditions.push(
			gte(transactions.date, `${year}-01-01`),
			lte(transactions.date, `${year}-12-31`),
		);
	}

	// Apply scope constraint from sharing
	if (access.type === "monthly") {
		conditions.push(
			gte(transactions.date, access.start),
			lte(transactions.date, access.end),
		);
	} else if (access.type === "yearly") {
		conditions.push(
			gte(transactions.date, access.start),
			lte(transactions.date, access.end),
		);
	}

	const where = and(...conditions);

	const [data, totals] = await Promise.all([
		db
			.select({
				id: transactions.id,
				date: transactions.date,
				description: transactions.description,
				amount: transactions.amount,
				type: transactions.type,
				categoryName: categories.name,
				notes: transactions.notes,
			})
			.from(transactions)
			.leftJoin(categories, eq(transactions.categoryId, categories.id))
			.where(where)
			.orderBy(desc(transactions.date), desc(transactions.id)),

		db
			.select({
				income: sql<number>`coalesce(sum(case when ${transactions.type} = 'income' then ${transactions.amount} else 0 end), 0)`,
				expense: sql<number>`coalesce(sum(case when ${transactions.type} = 'expense' then ${transactions.amount} else 0 end), 0)`,
				count: sql<number>`count(*)`,
			})
			.from(transactions)
			.where(where),
	]);

	return NextResponse.json({
		data,
		totals: {
			income: Number(totals[0].income),
			expense: Number(totals[0].expense),
			balance: Number(totals[0].income) - Number(totals[0].expense),
			count: Number(totals[0].count),
		},
	});
}
