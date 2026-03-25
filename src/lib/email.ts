import Email from "email-templates";
import nodemailer from "nodemailer";
import path from "node:path";

export interface BillEmailData {
	name: string;
	overdueBills: Array<{
		name: string;
		dueDay: number;
		daysOverdue: number;
		amountFormatted: string;
	}>;
	upcomingBills: Array<{
		name: string;
		dueDay: number;
		daysUntil: number;
		amountFormatted: string;
	}>;
	appUrl: string;
	settingsUrl: string;
	subject?: string;
}

const transport = nodemailer.createTransport({
	service: "icloud",
	secure: false,
	auth: {
		user: process.env.MAIL_USER,
		pass: process.env.MAIL_PASS,
	},
});

export async function sendBillNotificationEmail(
	to: string,
	data: BillEmailData,
) {
	const email = new Email({
		message: {
			from: `"Reconta" <${process.env.MAIL_ADDR!}>`,
		},
		transport,
		views: {
			root: path.join(process.cwd(), "src", "emails"),
			options: { extension: "pug" },
		},
		send: true,
		preview: false,
	});

	await email.send({
		template: "contas-vencendo",
		message: { to },
		locals: data,
	});
}
