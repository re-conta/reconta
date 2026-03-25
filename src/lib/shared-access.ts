import { and, eq } from "drizzle-orm";
import { db } from "@/lib/db";
import { sharedAccess } from "@/lib/db/schema";
import { getMonthRange } from "@/lib/utils";

export type ScopeConstraint =
	| { type: "all" }
	| { type: "yearly"; start: string; end: string }
	| { type: "monthly"; start: string; end: string };

/**
 * Checks if `viewerId` has access to `ownerId`'s data.
 * Returns null if no access, or the constraint to apply when querying.
 */
export async function checkSharedAccess(
	ownerId: string,
	viewerId: string,
): Promise<ScopeConstraint | null> {
	const rows = await db
		.select()
		.from(sharedAccess)
		.where(
			and(
				eq(sharedAccess.ownerId, ownerId),
				eq(sharedAccess.targetId, viewerId),
			),
		);

	if (rows.length === 0) return null;

	// If any share is "all", grant full access
	const allShare = rows.find((r) => r.scope === "all");
	if (allShare) return { type: "all" };

	// Prefer the widest scope available. Collect all allowed year/month ranges.
	// For simplicity, return the first yearly scope found, otherwise first monthly.
	const yearlyShare = rows.find((r) => r.scope === "yearly" && r.scopeYear);
	if (yearlyShare) {
		const y = yearlyShare.scopeYear!;
		return {
			type: "yearly",
			start: `${y}-01-01`,
			end: `${y}-12-31`,
		};
	}

	const monthlyShare = rows.find(
		(r) => r.scope === "monthly" && r.scopeMonth && r.scopeYear,
	);
	if (monthlyShare) {
		const { start, end } = getMonthRange(
			monthlyShare.scopeMonth!,
			monthlyShare.scopeYear!,
		);
		return { type: "monthly", start, end };
	}

	return null;
}

/**
 * Returns all active shares that `viewerId` has from `ownerId`.
 * Used when viewer wants the list of what months/years they can see.
 */
export async function getShares(ownerId: string, viewerId: string) {
	return db
		.select()
		.from(sharedAccess)
		.where(
			and(
				eq(sharedAccess.ownerId, ownerId),
				eq(sharedAccess.targetId, viewerId),
			),
		);
}
