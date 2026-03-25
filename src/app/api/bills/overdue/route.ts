export const dynamic = "force-dynamic";

import { and, eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { billPayments, bills } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function GET() {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const today = new Date();
	const currentMonth = today.getMonth() + 1;
	const currentYear = today.getFullYear();
	const todayDay = today.getDate();
	const daysInMonth = new Date(currentYear, currentMonth, 0).getDate();

	const data = await db
		.select({
			id: bills.id,
			name: bills.name,
			amount: bills.amount,
			dueDay: bills.dueDay,
			isPaid: billPayments.isPaid,
		})
		.from(bills)
		.leftJoin(
			billPayments,
			and(
				eq(billPayments.billId, bills.id),
				eq(billPayments.month, currentMonth),
				eq(billPayments.year, currentYear),
			),
		)
		.where(and(eq(bills.userId, userId), eq(bills.isActive, true)));

	const result = data
		.filter((b) => !b.isPaid)
		.map((b) => {
			const effectiveDueDay = Math.min(b.dueDay, daysInMonth);
			const daysUntil = effectiveDueDay - todayDay;
			return {
				id: b.id,
				name: b.name,
				amount: b.amount,
				dueDay: effectiveDueDay,
				daysUntil,
				isOverdue: daysUntil < 0,
				daysOverdue: daysUntil < 0 ? Math.abs(daysUntil) : 0,
			};
		})
		// Only bills that are overdue or due within 7 days
		.filter((b) => b.isOverdue || b.daysUntil <= 7)
		.sort((a, b) => a.daysUntil - b.daysUntil);

	return NextResponse.json(result);
}
