import { type NextRequest, NextResponse } from "next/server";
import { sendTextMessage, sendTemplateMessage } from "@/lib/whatsapp";

export async function POST(req: NextRequest) {
	try {
		const { to, message, type = "text", template } = await req.json();

		if (!to) {
			return NextResponse.json(
				{ error: "Campo 'to' é obrigatório" },
				{ status: 400 },
			);
		}

		let result: unknown;

		if (type === "template") {
			result = await sendTemplateMessage(to, template ?? "hello_world");
		} else {
			if (!message) {
				return NextResponse.json(
					{ error: "Campo 'message' é obrigatório para tipo 'text'" },
					{ status: 400 },
				);
			}
			result = await sendTextMessage(to, message);
		}

		return NextResponse.json({ success: true, data: result });
	} catch (error) {
		console.error("Erro ao enviar mensagem:", error);
		return NextResponse.json(
			{ error: error instanceof Error ? error.message : "Erro desconhecido" },
			{ status: 500 },
		);
	}
}
