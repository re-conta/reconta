export const dynamic = "force-dynamic";

import { eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { user } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function DELETE() {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	await db.delete(user).where(eq(user.id, userId));

	return NextResponse.json({ ok: true });
}

export async function PATCH(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const body = await request.json();
	const { name, image } = body as { name?: string; image?: string | null };

	const updates: Record<string, unknown> = { updatedAt: new Date() };
	if (name !== undefined) updates.name = String(name).trim();
	if (image !== undefined) updates.image = image || null;

	const [updated] = await db
		.update(user)
		.set(updates)
		.where(eq(user.id, userId))
		.returning({ name: user.name, image: user.image });

	return NextResponse.json(updated);
}
