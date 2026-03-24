export const dynamic = "force-dynamic";

import { eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { accounts } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function GET() {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const data = await db
		.select()
		.from(accounts)
		.where(eq(accounts.userId, userId))
		.orderBy(accounts.name);

	return NextResponse.json(data);
}

export async function POST(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const body = await request.json();
	const { name, type, balance } = body;

	if (!name) {
		return NextResponse.json({ error: "Nome é obrigatório" }, { status: 400 });
	}

	const [account] = await db
		.insert(accounts)
		.values({
			userId,
			name,
			type: type ?? "checking",
			balance: Number(balance ?? 0),
		})
		.returning();

	return NextResponse.json(account, { status: 201 });
}
