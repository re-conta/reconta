export const dynamic = "force-dynamic";

import { and, eq } from "drizzle-orm";
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
			targetId: sharedAccess.targetId,
			targetName: user.name,
			targetEmail: user.email,
			scope: sharedAccess.scope,
			scopeMonth: sharedAccess.scopeMonth,
			scopeYear: sharedAccess.scopeYear,
			createdAt: sharedAccess.createdAt,
		})
		.from(sharedAccess)
		.innerJoin(user, eq(sharedAccess.targetId, user.id))
		.where(eq(sharedAccess.ownerId, userId));

	return NextResponse.json(rows);
}

export async function POST(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const body = await request.json();
	const { email, scope, scopeMonth, scopeYear } = body as {
		email: string;
		scope: "all" | "yearly" | "monthly";
		scopeMonth?: number;
		scopeYear?: number;
	};

	if (!email || !scope) {
		return NextResponse.json(
			{ error: "Email e escopo são obrigatórios" },
			{ status: 400 },
		);
	}

	const scopeValues = ["all", "yearly", "monthly"];
	if (!scopeValues.includes(scope)) {
		return NextResponse.json({ error: "Escopo inválido" }, { status: 400 });
	}

	if (scope === "monthly" && (!scopeMonth || !scopeYear)) {
		return NextResponse.json(
			{ error: "Mês e ano são obrigatórios para escopo mensal" },
			{ status: 400 },
		);
	}

	if (scope === "yearly" && !scopeYear) {
		return NextResponse.json(
			{ error: "Ano é obrigatório para escopo anual" },
			{ status: 400 },
		);
	}

	const [targetUser] = await db
		.select({ id: user.id, name: user.name, email: user.email })
		.from(user)
		.where(eq(user.email, email.toLowerCase().trim()));

	if (!targetUser) {
		return NextResponse.json(
			{ error: "Usuário não encontrado com esse e-mail" },
			{ status: 404 },
		);
	}

	if (targetUser.id === userId) {
		return NextResponse.json(
			{ error: "Você não pode compartilhar com você mesmo" },
			{ status: 400 },
		);
	}

	const [existing] = await db
		.select({ id: sharedAccess.id })
		.from(sharedAccess)
		.where(
			and(
				eq(sharedAccess.ownerId, userId),
				eq(sharedAccess.targetId, targetUser.id),
				eq(sharedAccess.scope, scope),
			),
		);

	if (existing) {
		return NextResponse.json(
			{ error: "Acesso já compartilhado com esse usuário neste escopo" },
			{ status: 409 },
		);
	}

	const [inserted] = await db
		.insert(sharedAccess)
		.values({
			ownerId: userId,
			targetId: targetUser.id,
			scope,
			scopeMonth: scope === "monthly" ? scopeMonth : null,
			scopeYear: scope === "monthly" || scope === "yearly" ? scopeYear : null,
		})
		.returning();

	return NextResponse.json(
		{
			id: inserted.id,
			targetName: targetUser.name,
			targetEmail: targetUser.email,
			scope: inserted.scope,
			scopeMonth: inserted.scopeMonth,
			scopeYear: inserted.scopeYear,
			createdAt: inserted.createdAt,
		},
		{ status: 201 },
	);
}
