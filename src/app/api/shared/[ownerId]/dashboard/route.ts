export const dynamic = "force-dynamic";

import { and, desc, eq, gte, lte, sql } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { categories, transactions } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";
import { checkSharedAccess } from "@/lib/shared-access";
import {
	getMonthRange,
	getPreviousMonth,
	getPreviousPeriod,
	getYearRange,
} from "@/lib/utils";

async function getTotal(
	type: "income" | "expense",
	start: string,
	end: string,
	userId: string,
) {
	const [row] = await db
		.select({ total: sql<number>`coalesce(sum(${transactions.amount}), 0)` })
		.from(transactions)
		.where(
			and(
				eq(transactions.userId, userId),
				eq(transactions.type, type),
				gte(transactions.date, start),
				lte(transactions.date, end),
			),
		);
	return Number(row?.total ?? 0);
}

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
	const scope = searchParams.get("scope") ?? "month";

	let start: string;
	let end: string;
	let hasComparison = true;
	let prevStart = "";
	let prevEnd = "";

	if (scope === "year") {
		({ start, end } = getYearRange(year));
		({ start: prevStart, end: prevEnd } = getYearRange(year - 1));
	} else if (scope === "all") {
		start = "1970-01-01";
		end = new Date().toISOString().split("T")[0];
		hasComparison = false;
	} else if (scope === "custom") {
		start = searchParams.get("start") ?? new Date().toISOString().split("T")[0];
		end = searchParams.get("end") ?? new Date().toISOString().split("T")[0];
		({ start: prevStart, end: prevEnd } = getPreviousPeriod(start, end));
	} else {
		({ start, end } = getMonthRange(month, year));
		const prev = getPreviousMonth(month, year);
		({ start: prevStart, end: prevEnd } = getMonthRange(prev.month, prev.year));
	}

	const [currentIncome, currentExpense, prevIncome, prevExpense] =
		await Promise.all([
			getTotal("income", start, end, ownerId),
			getTotal("expense", start, end, ownerId),
			hasComparison
				? getTotal("income", prevStart, prevEnd, ownerId)
				: Promise.resolve(0),
			hasComparison
				? getTotal("expense", prevStart, prevEnd, ownerId)
				: Promise.resolve(0),
		]);

	const expensesByCategory = await db
		.select({
			categoryId: transactions.categoryId,
			categoryName: categories.name,
			categoryColor: categories.color,
			total: sql<number>`coalesce(sum(${transactions.amount}), 0)`,
		})
		.from(transactions)
		.leftJoin(categories, eq(transactions.categoryId, categories.id))
		.where(
			and(
				eq(transactions.userId, ownerId),
				eq(transactions.type, "expense"),
				gte(transactions.date, start),
				lte(transactions.date, end),
			),
		)
		.groupBy(transactions.categoryId, categories.name, categories.color)
		.orderBy(sql`coalesce(sum(${transactions.amount}), 0) desc`);

	const recentTransactions = await db
		.select({
			id: transactions.id,
			date: transactions.date,
			description: transactions.description,
			amount: transactions.amount,
			type: transactions.type,
			categoryName: categories.name,
			categoryColor: categories.color,
		})
		.from(transactions)
		.leftJoin(categories, eq(transactions.categoryId, categories.id))
		.where(
			and(
				eq(transactions.userId, ownerId),
				gte(transactions.date, start),
				lte(transactions.date, end),
			),
		)
		.orderBy(desc(transactions.date))
		.limit(10);

	const trendEnd = new Date(`${end}T00:00:00`);
	const trendMonth = trendEnd.getMonth() + 1;
	const trendYear = trendEnd.getFullYear();

	const monthlyBalance = [];
	for (let i = 5; i >= 0; i--) {
		let m = trendMonth - i;
		let y = trendYear;
		while (m <= 0) {
			m += 12;
			y -= 1;
		}
		const { start: ms, end: me } = getMonthRange(m, y);
		const [inc, exp] = await Promise.all([
			getTotal("income", ms, me, ownerId),
			getTotal("expense", ms, me, ownerId),
		]);
		monthlyBalance.push({
			month: m,
			year: y,
			income: inc,
			expense: exp,
			balance: inc - exp,
		});
	}

	return NextResponse.json({
		current: {
			income: currentIncome,
			expense: currentExpense,
			balance: currentIncome - currentExpense,
		},
		previous: {
			income: prevIncome,
			expense: prevExpense,
			balance: prevIncome - prevExpense,
		},
		hasComparison,
		expensesByCategory,
		recentTransactions,
		monthlyBalance,
	});
}
