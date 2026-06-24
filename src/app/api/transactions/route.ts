export const dynamic = "force-dynamic";

import { and, desc, eq, gte, inArray, like, lte, sql } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import {
	categories,
	tags,
	transactionTags,
	transactions,
} from "@/lib/db/schema";
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
	const tagId = searchParams.get("tagId");
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

	if (tagId) {
		const taggedRows = await db
			.select({ id: transactionTags.transactionId })
			.from(transactionTags)
			.where(eq(transactionTags.tagId, Number(tagId)));
		conditions.push(
			inArray(transactions.id, taggedRows.map((r) => r.id).concat(-1)),
		);
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
				bank: transactions.bank,
				pixBeneficiary: transactions.pixBeneficiary,
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

	const txIds = data.map((t) => t.id);
	const tagRows = txIds.length
		? await db
				.select({
					transactionId: transactionTags.transactionId,
					id: tags.id,
					name: tags.name,
					color: tags.color,
				})
				.from(transactionTags)
				.innerJoin(tags, eq(transactionTags.tagId, tags.id))
				.where(inArray(transactionTags.transactionId, txIds))
		: [];

	const tagsByTx = new Map<
		number,
		{ id: number; name: string; color: string }[]
	>();
	for (const row of tagRows) {
		const list = tagsByTx.get(row.transactionId) ?? [];
		list.push({ id: row.id, name: row.name, color: row.color });
		tagsByTx.set(row.transactionId, list);
	}

	const dataWithTags = data.map((t) => ({
		...t,
		tags: tagsByTx.get(t.id) ?? [],
	}));

	return NextResponse.json({
		data: dataWithTags,
		totals: {
			income: Number(totals[0].income),
			expense: Number(totals[0].expense),
			balance: Number(totals[0].income) - Number(totals[0].expense),
			count: Number(totals[0].count),
		},
		pagination: { page, limit, total: Number(totals[0].count) },
	});
}

export async function DELETE(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const body = await request.json();
	const { scope, month, year } = body;

	if (!["month", "year", "all"].includes(scope)) {
		return NextResponse.json({ error: "Escopo inválido" }, { status: 400 });
	}

	const conditions = [eq(transactions.userId, userId)];

	if (scope === "month") {
		const { start, end } = getMonthRange(Number(month), Number(year));
		conditions.push(gte(transactions.date, start), lte(transactions.date, end));
	} else if (scope === "year") {
		conditions.push(
			gte(transactions.date, `${year}-01-01`),
			lte(transactions.date, `${year}-12-31`),
		);
	}

	const deleted = await db
		.delete(transactions)
		.where(and(...conditions))
		.returning({ id: transactions.id });

	return NextResponse.json({ deleted: deleted.length });
}

export async function PATCH(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const body = await request.json();
	const { ids, fields } = body;

	if (!Array.isArray(ids) || ids.length === 0) {
		return NextResponse.json({ error: "IDs obrigatórios" }, { status: 400 });
	}

	const update: {
		type?: "income" | "expense";
		categoryId?: number | null;
		accountId?: number | null;
		date?: string;
	} = {};

	if (fields.type !== undefined) update.type = fields.type;
	if ("categoryId" in fields)
		update.categoryId =
			fields.categoryId && fields.categoryId !== "_none"
				? Number(fields.categoryId)
				: null;
	if ("accountId" in fields)
		update.accountId =
			fields.accountId && fields.accountId !== "_none"
				? Number(fields.accountId)
				: null;
	if (fields.date !== undefined) update.date = fields.date;

	if (Object.keys(update).length === 0) {
		return NextResponse.json({ error: "Nenhum campo" }, { status: 400 });
	}

	const updated = await db
		.update(transactions)
		.set(update)
		.where(and(eq(transactions.userId, userId), inArray(transactions.id, ids)))
		.returning({ id: transactions.id });

	return NextResponse.json({ updated: updated.length });
}

export async function POST(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const body = await request.json();
	const {
		date,
		description,
		amount,
		type,
		categoryId,
		accountId,
		notes,
		tagIds,
	} = body;

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

	if (Array.isArray(tagIds) && tagIds.length > 0) {
		await db.insert(transactionTags).values(
			tagIds.map((tagId: number) => ({
				transactionId: transaction.id,
				tagId: Number(tagId),
			})),
		);
	}

	return NextResponse.json(transaction, { status: 201 });
}
