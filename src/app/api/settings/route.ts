export const dynamic = "force-dynamic";

import { eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { notificationSettings } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";

export async function GET() {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const [settings] = await db
		.select()
		.from(notificationSettings)
		.where(eq(notificationSettings.userId, userId));

	if (!settings) {
		// Return defaults without persisting
		return NextResponse.json({
			enabled: true,
			emailAddress: null,
			daysBeforeDue: 3,
			daysAfterDue: 7,
			maxNotificationsPerBill: 3,
			intervalDays: 1,
		});
	}

	return NextResponse.json(settings);
}

export async function PUT(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	const body = await request.json();
	const {
		enabled,
		emailAddress,
		daysBeforeDue,
		daysAfterDue,
		maxNotificationsPerBill,
		intervalDays,
	} = body;

	const now = new Date().toISOString().replace("T", " ").slice(0, 19);

	const [existing] = await db
		.select({ id: notificationSettings.id })
		.from(notificationSettings)
		.where(eq(notificationSettings.userId, userId));

	if (existing) {
		const [updated] = await db
			.update(notificationSettings)
			.set({
				enabled: Boolean(enabled),
				emailAddress: emailAddress || null,
				daysBeforeDue: Number(daysBeforeDue),
				daysAfterDue: Number(daysAfterDue),
				maxNotificationsPerBill: Number(maxNotificationsPerBill),
				intervalDays: Number(intervalDays),
				updatedAt: now,
			})
			.where(eq(notificationSettings.userId, userId))
			.returning();
		return NextResponse.json(updated);
	}

	const [created] = await db
		.insert(notificationSettings)
		.values({
			userId,
			enabled: Boolean(enabled),
			emailAddress: emailAddress || null,
			daysBeforeDue: Number(daysBeforeDue),
			daysAfterDue: Number(daysAfterDue),
			maxNotificationsPerBill: Number(maxNotificationsPerBill),
			intervalDays: Number(intervalDays),
			updatedAt: now,
		})
		.returning();

	return NextResponse.json(created);
}
