export const dynamic = "force-dynamic";

import { and, eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { bills } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function PUT(
	request: Request,
	{ params }: { params: Promise<{ id: string }> },
) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { id } = await params;
	const body = await request.json();
	const { name, amount, dueDay, categoryId, isActive } = body;

	const [updated] = await db
		.update(bills)
		.set({
			name,
			amount: Number(amount),
			dueDay: Number(dueDay),
			categoryId: categoryId ? Number(categoryId) : null,
			isActive: isActive ?? true,
		})
		.where(and(eq(bills.id, Number(id)), eq(bills.userId, userId)))
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
		.delete(bills)
		.where(and(eq(bills.id, Number(id)), eq(bills.userId, userId)));

	return NextResponse.json({ success: true });
}
