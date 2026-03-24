export const dynamic = "force-dynamic";

import { and, desc, eq, gte, like, lte, sql } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { categories, transactions } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";
import { getMonthRange } from "@/lib/utils";

export async function GET(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { searchParams } = new URL(request.url);
	const month = searchParams.get("month");
	const year = searchParams.get("year");
	const type = searchParams.get("type");
	const categoryId = searchParams.get("categoryId");
	const search = searchParams.get("search");
	const page = Number(searchParams.get("page") ?? 1);
	const limit = Number(searchParams.get("limit") ?? 50);
	const offset = (page - 1) * limit;

	const conditions = [eq(transactions.userId, userId)];

	if (month && year) {
		const { start, end } = getMonthRange(Number(month), Number(year));
		conditions.push(gte(transactions.date, start), lte(transactions.date, end));
	}

	if (type && (type === "income" || type === "expense")) {
		conditions.push(eq(transactions.type, type));
	}

	if (categoryId) {
		conditions.push(eq(transactions.categoryId, Number(categoryId)));
	}

	if (search) {
		conditions.push(like(transactions.description, `%${search}%`));
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
				categoryId: transactions.categoryId,
				categoryName: categories.name,
				categoryColor: categories.color,
				accountId: transactions.accountId,
				notes: transactions.notes,
				importedFrom: transactions.importedFrom,
				createdAt: transactions.createdAt,
			})
			.from(transactions)
			.leftJoin(categories, eq(transactions.categoryId, categories.id))
			.where(where)
			.orderBy(desc(transactions.date), desc(transactions.id))
			.limit(limit)
			.offset(offset),

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
		pagination: { page, limit, total: Number(totals[0].count) },
	});
}

export async function POST(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const body = await request.json();
	const { date, description, amount, type, categoryId, accountId, notes } =
		body;

	if (!date || !description || !amount || !type) {
		return NextResponse.json(
			{ error: "Campos obrigatórios faltando" },
			{ status: 400 },
		);
	}

	const [transaction] = await db
		.insert(transactions)
		.values({
			userId,
			date,
			description,
			amount: Math.abs(Number(amount)),
			type,
			categoryId: categoryId ? Number(categoryId) : null,
			accountId: accountId ? Number(accountId) : null,
			notes: notes || null,
		})
		.returning();

	return NextResponse.json(transaction, { status: 201 });
}
