export const dynamic = "force-dynamic";

import { and, eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { billPayments, bills } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function POST(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const body = await request.json();
	const { billId, month, year, isPaid, amount } = body;

	// Verify the bill belongs to this user
	const [bill] = await db
		.select({ id: bills.id })
		.from(bills)
		.where(and(eq(bills.id, Number(billId)), eq(bills.userId, userId)))
		.limit(1);

	if (!bill) {
		return NextResponse.json({ error: "Não autorizado" }, { status: 403 });
	}

	const existing = await db
		.select()
		.from(billPayments)
		.where(
			and(
				eq(billPayments.billId, Number(billId)),
				eq(billPayments.month, Number(month)),
				eq(billPayments.year, Number(year)),
			),
		)
		.limit(1);

	if (existing.length > 0) {
		const [updated] = await db
			.update(billPayments)
			.set({
				isPaid: Boolean(isPaid),
				paidAt: isPaid ? new Date().toISOString() : null,
				amount: amount ? Number(amount) : null,
			})
			.where(eq(billPayments.id, existing[0].id))
			.returning();
		return NextResponse.json(updated);
	}

	const [payment] = await db
		.insert(billPayments)
		.values({
			billId: Number(billId),
			month: Number(month),
			year: Number(year),
			isPaid: Boolean(isPaid),
			paidAt: isPaid ? new Date().toISOString() : null,
			amount: amount ? Number(amount) : null,
		})
		.returning();

	return NextResponse.json(payment, { status: 201 });
}
