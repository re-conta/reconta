export const dynamic = "force-dynamic";

import { eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { categories } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";
import { checkSharedAccess } from "@/lib/shared-access";

export async function GET(
	_request: Request,
	{ params }: { params: Promise<{ ownerId: string }> },
) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { ownerId } = await params;

	const access = await checkSharedAccess(ownerId, userId);
	if (!access) {
		return NextResponse.json({ error: "Acesso negado" }, { status: 403 });
	}

	const data = await db
		.select()
		.from(categories)
		.where(eq(categories.userId, ownerId))
		.orderBy(categories.name);

	return NextResponse.json(data);
}
