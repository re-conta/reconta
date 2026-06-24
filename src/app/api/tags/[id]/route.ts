export const dynamic = "force-dynamic";

import { and, eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { tags } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function PUT(
	request: Request,
	{ params }: { params: Promise<{ id: string }> },
) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { id } = await params;
	const body = await request.json();
	const { name, color } = body;

	const [updated] = await db
		.update(tags)
		.set({ name, color })
		.where(and(eq(tags.id, Number(id)), eq(tags.userId, userId)))
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
		.delete(tags)
		.where(and(eq(tags.id, Number(id)), eq(tags.userId, userId)));

	return NextResponse.json({ success: true });
}
