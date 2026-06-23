export const dynamic = "force-dynamic";

import { and, eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { categories } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function PUT(
	request: Request,
	{ params }: { params: Promise<{ id: string }> },
) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { id } = await params;
	const body = await request.json();
	const { name, color, icon, type, patterns } = body;

	const [updated] = await db
		.update(categories)
		.set({ name, color, icon, type, patterns: patterns || null })
		.where(and(eq(categories.id, Number(id)), eq(categories.userId, userId)))
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
		.delete(categories)
		.where(and(eq(categories.id, Number(id)), eq(categories.userId, userId)));

	return NextResponse.json({ success: true });
}
