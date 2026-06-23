export const dynamic = "force-dynamic";

import { eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { categories } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function GET() {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const data = await db
		.select()
		.from(categories)
		.where(eq(categories.userId, userId))
		.orderBy(categories.name);

	return NextResponse.json(data);
}

export async function POST(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const body = await request.json();
	const { name, color, icon, type, patterns } = body;

	if (!name) {
		return NextResponse.json({ error: "Nome é obrigatório" }, { status: 400 });
	}

	const [category] = await db
		.insert(categories)
		.values({
			userId,
			name,
			color: color ?? "#6366f1",
			icon: icon ?? "circle",
			type: type ?? "both",
			patterns: patterns || null,
		})
		.returning();

	return NextResponse.json(category, { status: 201 });
}
