export const dynamic = "force-dynamic";

import { and, eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { transactions } from "@/lib/db/schema";
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
	return NextResponse.json(tx);
}

export async function PUT(
	request: Request,
	{ params }: { params: Promise<{ id: string }> },
) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { id } = await params;
	const body = await request.json();
	const { date, description, amount, type, categoryId, accountId, notes } =
		body;

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
