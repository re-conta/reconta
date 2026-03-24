export const dynamic = "force-dynamic";

import { and, eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { accounts } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function PUT(
	request: Request,
	{ params }: { params: Promise<{ id: string }> },
) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { id } = await params;
	const body = await request.json();
	const { name, type, balance } = body;

	const [updated] = await db
		.update(accounts)
		.set({ name, type, balance: Number(balance) })
		.where(and(eq(accounts.id, Number(id)), eq(accounts.userId, userId)))
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
		.delete(accounts)
		.where(and(eq(accounts.id, Number(id)), eq(accounts.userId, userId)));

	return NextResponse.json({ success: true });
}
