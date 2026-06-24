export const dynamic = "force-dynamic";

import { eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { tags } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function GET() {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const data = await db
		.select()
		.from(tags)
		.where(eq(tags.userId, userId))
		.orderBy(tags.name);

	return NextResponse.json(data);
}

export async function POST(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const body = await request.json();
	const { name, color } = body;

	if (!name) {
		return NextResponse.json({ error: "Nome é obrigatório" }, { status: 400 });
	}

	const [tag] = await db
		.insert(tags)
		.values({
			userId,
			name: name.trim(),
			color: color ?? "#6366f1",
		})
		.returning();

	return NextResponse.json(tag, { status: 201 });
}
