export const dynamic = "force-dynamic";

import { and, eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { tags, transactionTags, transactions } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function GET(
	_request: Request,
	{ params }: { params: Promise<{ id: string }> },
) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { id } = await params;
	const [tx] = await db
		.select()
		.from(transactions)
		.where(
			and(eq(transactions.id, Number(id)), eq(transactions.userId, userId)),
		);

	if (!tx)
		return NextResponse.json({ error: "Não encontrado" }, { status: 404 });

	const txTags = await db
		.select({ id: tags.id, name: tags.name, color: tags.color })
		.from(transactionTags)
		.innerJoin(tags, eq(transactionTags.tagId, tags.id))
		.where(eq(transactionTags.transactionId, tx.id));

	return NextResponse.json({ ...tx, tags: txTags });
}

export async function PUT(
	request: Request,
	{ params }: { params: Promise<{ id: string }> },
) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { id } = await params;
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

	const [updated] = await db
		.update(transactions)
		.set({
			date,
			description,
			amount: Math.abs(Number(amount)),
			type,
			categoryId: categoryId ? Number(categoryId) : null,
			accountId: accountId ? Number(accountId) : null,
			notes: notes || null,
		})
		.where(
			and(eq(transactions.id, Number(id)), eq(transactions.userId, userId)),
		)
		.returning();

	if (!updated)
		return NextResponse.json({ error: "Não encontrado" }, { status: 404 });

	if (Array.isArray(tagIds)) {
		await db
			.delete(transactionTags)
			.where(eq(transactionTags.transactionId, updated.id));
		if (tagIds.length > 0) {
			await db.insert(transactionTags).values(
				tagIds.map((tagId: number) => ({
					transactionId: updated.id,
					tagId: Number(tagId),
				})),
			);
		}
	}

	return NextResponse.json(updated);
}

export async function DELETE(
	_request: Request,
	{ params }: { params: Promise<{ id: string }> },
) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { id } = await params;
	await db
		.delete(transactions)
		.where(
			and(eq(transactions.id, Number(id)), eq(transactions.userId, userId)),
		);

	return NextResponse.json({ success: true });
}
