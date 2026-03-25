export const dynamic = "force-dynamic";

import { eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { sharedAccess, user } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function GET() {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const rows = await db
		.select({
			id: sharedAccess.id,
			ownerId: sharedAccess.ownerId,
			ownerName: user.name,
			ownerEmail: user.email,
			scope: sharedAccess.scope,
			scopeMonth: sharedAccess.scopeMonth,
			scopeYear: sharedAccess.scopeYear,
			createdAt: sharedAccess.createdAt,
		})
		.from(sharedAccess)
		.innerJoin(user, eq(sharedAccess.ownerId, user.id))
		.where(eq(sharedAccess.targetId, userId));

	return NextResponse.json(rows);
}
