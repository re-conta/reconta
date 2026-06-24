export const dynamic = "force-dynamic";

import { and, eq, isNull, ne } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { categories, transactions } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function POST() {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const userCategories = await db
		.select({ id: categories.id, patterns: categories.patterns })
		.from(categories)
		.where(and(eq(categories.userId, userId), ne(categories.patterns, "")));

	const rules = userCategories
		.map((c) => ({
			id: c.id,
			patterns: (c.patterns ?? "")
				.split("\n")
				.map((p) => p.trim())
				.filter(Boolean)
				.map((p) => {
					try {
						return new RegExp(p, "i");
					} catch {
						return new RegExp(p.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"), "i");
					}
				}),
		}))
		.filter((c) => c.patterns.length > 0);

	if (rules.length === 0) {
		return NextResponse.json({ updated: 0, checked: 0 });
	}

	const uncategorized = await db
		.select({
			id: transactions.id,
			description: transactions.description,
			pixBeneficiary: transactions.pixBeneficiary,
		})
		.from(transactions)
		.where(
			and(eq(transactions.userId, userId), isNull(transactions.categoryId)),
		);

	const updates: { id: number; categoryId: number }[] = [];
	for (const tx of uncategorized) {
		const haystack = `${tx.description} ${tx.pixBeneficiary ?? ""}`;
		const match = rules.find((rule) =>
			rule.patterns.some((pattern) => pattern.test(haystack)),
		);
		if (match) updates.push({ id: tx.id, categoryId: match.id });
	}

	if (updates.length > 0) {
		db.transaction((tx) => {
			for (const u of updates) {
				tx.update(transactions)
					.set({ categoryId: u.categoryId })
					.where(
						and(eq(transactions.id, u.id), eq(transactions.userId, userId)),
					)
					.run();
			}
		});
	}

	return NextResponse.json({
		updated: updates.length,
		checked: uncategorized.length,
	});
}
