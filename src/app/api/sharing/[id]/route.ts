export const dynamic = "force-dynamic";

import { and, eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { sharedAccess } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function DELETE(
	_request: Request,
	{ params }: { params: Promise<{ id: string }> },
) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { id } = await params;
	const shareId = Number(id);

	if (Number.isNaN(shareId)) {
		return NextResponse.json({ error: "ID inválido" }, { status: 400 });
	}

	const result = await db
		.delete(sharedAccess)
		.where(
			and(eq(sharedAccess.id, shareId), eq(sharedAccess.ownerId, userId)),
		)
		.returning({ id: sharedAccess.id });

	if (result.length === 0) {
		return NextResponse.json(
			{ error: "Compartilhamento não encontrado" },
			{ status: 404 },
		);
	}

	return NextResponse.json({ success: true });
}
