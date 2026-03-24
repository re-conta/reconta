export const dynamic = "force-dynamic";

import { and, asc, eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { billPayments, bills, categories } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function GET(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const { searchParams } = new URL(request.url);
	const month = Number(searchParams.get("month") ?? new Date().getMonth() + 1);
	const year = Number(searchParams.get("year") ?? new Date().getFullYear());

	const data = await db
		.select({
			id: bills.id,
			name: bills.name,
			amount: bills.amount,
			dueDay: bills.dueDay,
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
		.where(and(eq(bills.userId, userId), eq(bills.isActive, true)))
		.orderBy(asc(bills.dueDay));

	return NextResponse.json(data);
}

export async function POST(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const body = await request.json();
	const { name, amount, dueDay, categoryId } = body;

	if (!name || !amount || !dueDay) {
		return NextResponse.json(
			{ error: "Campos obrigatórios faltando" },
			{ status: 400 },
		);
	}

	const [bill] = await db
		.insert(bills)
		.values({
			userId,
			name,
			amount: Number(amount),
			dueDay: Number(dueDay),
			categoryId: categoryId ? Number(categoryId) : null,
		})
		.returning();

	return NextResponse.json(bill, { status: 201 });
}
