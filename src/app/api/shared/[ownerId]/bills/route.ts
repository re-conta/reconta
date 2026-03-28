export const dynamic = "force-dynamic";

import { and, asc, eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { billPayments, bills, categories } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";
import { checkSharedAccess } from "@/lib/shared-access";

export async function GET(
	request: Request,
	{ params }: { params: Promise<{ ownerId: string }> },
) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { ownerId } = await params;

	const access = await checkSharedAccess(ownerId, userId);
	if (!access) {
		return NextResponse.json({ error: "Acesso negado" }, { status: 403 });
	}

	const { searchParams } = new URL(request.url);
	const month = Number(searchParams.get("month") ?? new Date().getMonth() + 1);
	const year = Number(searchParams.get("year") ?? new Date().getFullYear());

	const data = await db
		.select({
			id: bills.id,
			name: bills.name,
			amount: bills.amount,
			dueDay: bills.dueDay,
			frequency: bills.frequency,
			isActive: bills.isActive,
			categoryId: bills.categoryId,
			categoryName: categories.name,
			categoryColor: categories.color,
			paymentId: billPayments.id,
			isPaid: billPayments.isPaid,
			paidAt: billPayments.paidAt,
			paymentAmount: billPayments.amount,
		})
		.from(bills)
		.leftJoin(categories, eq(bills.categoryId, categories.id))
		.leftJoin(
			billPayments,
			and(
				eq(billPayments.billId, bills.id),
				eq(billPayments.month, month),
				eq(billPayments.year, year),
			),
		)
		.where(and(eq(bills.userId, ownerId), eq(bills.isActive, true)))
		.orderBy(asc(bills.dueDay));

	return NextResponse.json(data);
}
