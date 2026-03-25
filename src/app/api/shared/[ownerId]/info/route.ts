export const dynamic = "force-dynamic";

import { and, eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { sharedAccess, user } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function GET(
	_request: Request,
	{ params }: { params: Promise<{ ownerId: string }> },
) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { ownerId } = await params;

	const rows = await db
		.select({
			id: sharedAccess.id,
			scope: sharedAccess.scope,
			scopeMonth: sharedAccess.scopeMonth,
			scopeYear: sharedAccess.scopeYear,
			ownerName: user.name,
			ownerEmail: user.email,
		})
		.from(sharedAccess)
		.innerJoin(user, eq(sharedAccess.ownerId, user.id))
		.where(
			and(
				eq(sharedAccess.ownerId, ownerId),
				eq(sharedAccess.targetId, userId),
			),
		);

	if (rows.length === 0) {
		return NextResponse.json({ error: "Acesso negado" }, { status: 403 });
	}

	return NextResponse.json({
		ownerName: rows[0].ownerName,
		ownerEmail: rows[0].ownerEmail,
		shares: rows.map((r) => ({
			id: r.id,
			scope: r.scope,
			scopeMonth: r.scopeMonth,
			scopeYear: r.scopeYear,
		})),
	});
}
