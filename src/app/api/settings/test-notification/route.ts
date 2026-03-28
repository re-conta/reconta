export const dynamic = "force-dynamic";

import { eq } from "drizzle-orm";
import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { notificationSettings, user } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";
import { sendBillNotificationEmail } from "@/lib/email";
import { sendTextMessage } from "@/lib/whatsapp";

const ADMIN_EMAIL = "sistematico@gmail.com";

const sampleData = {
	overdueBills: [
		{
			name: "Internet Fibra",
			dueDay: 15,
			daysOverdue: 5,
			amountFormatted: "R$ 149,90",
		},
	],
	upcomingBills: [
		{
			name: "Aluguel",
			dueDay: 1,
			daysUntil: 3,
			amountFormatted: "R$ 2.500,00",
		},
		{
			name: "Energia Elétrica",
			dueDay: 2,
			daysUntil: 4,
			amountFormatted: "R$ 230,00",
		},
	],
};

export async function POST(request: Request) {
	const { userId, unauthorized } = await requireSession();
	if (unauthorized) return unauthorized;

	// Only allow admin user
	const [currentUser] = await db
		.select({ email: user.email })
		.from(user)
		.where(eq(user.id, userId));

	if (currentUser?.email !== ADMIN_EMAIL) {
		return NextResponse.json({ error: "Forbidden" }, { status: 403 });
	}

	const { channel } = await request.json();

	const [settings] = await db
		.select()
		.from(notificationSettings)
		.where(eq(notificationSettings.userId, userId));

	const appUrl = process.env.NEXT_PUBLIC_APP_URL ?? "https://reconta.app";
	const settingsUrl = `${appUrl}/ajustes`;

	if (channel === "email") {
		const emailTo =
			settings?.emailAddress || currentUser.email;
		await sendBillNotificationEmail(emailTo, {
			name: "Teste",
			...sampleData,
			appUrl,
			settingsUrl,
		});
		return NextResponse.json({ ok: true, sentTo: emailTo });
	}

	if (channel === "whatsapp") {
		const phone = settings?.whatsappNumber;
		if (!phone) {
			return NextResponse.json(
				{ error: "Número de WhatsApp não configurado" },
				{ status: 400 },
			);
		}

		const lines: string[] = [];
		lines.push("⚠️ *Alerta de contas — ReConta (TESTE)*");
		lines.push("");
		lines.push("🔴 *Contas vencidas:*");
		for (const b of sampleData.overdueBills) {
			lines.push(
				`• ${b.name} — dia ${b.dueDay} (${b.daysOverdue} dia${b.daysOverdue === 1 ? "" : "s"} atrás) — ${b.amountFormatted}`,
			);
		}
		lines.push("");
		lines.push("🟡 *Contas a vencer:*");
		for (const b of sampleData.upcomingBills) {
			lines.push(
				`• ${b.name} — dia ${b.dueDay} (em ${b.daysUntil} dia${b.daysUntil === 1 ? "" : "s"}) — ${b.amountFormatted}`,
			);
		}
		lines.push("");
		lines.push(`Acesse: ${appUrl}/contas`);

		await sendTextMessage(phone, lines.join("\n"));
		return NextResponse.json({ ok: true, sentTo: phone });
	}

	return NextResponse.json({ error: "Canal inválido" }, { status: 400 });
}
