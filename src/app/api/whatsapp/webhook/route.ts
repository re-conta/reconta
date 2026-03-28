import { type NextRequest, NextResponse } from "next/server";
import crypto from "node:crypto";

const VERIFY_TOKEN = process.env.WHATSAPP_VERIFY_TOKEN!;
const APP_SECRET = process.env.WHATSAPP_APP_SECRET;

// GET — Verificação do webhook pela Meta
export async function GET(req: NextRequest) {
	const searchParams = req.nextUrl.searchParams;
	const mode = searchParams.get("hub.mode");
	const token = searchParams.get("hub.verify_token");
	const challenge = searchParams.get("hub.challenge");

	if (mode === "subscribe" && token === VERIFY_TOKEN) {
		console.log("✅ Webhook verificado com sucesso");
		return new NextResponse(challenge, { status: 200 });
	}

	console.warn("❌ Falha na verificação do webhook");
	return NextResponse.json({ error: "Verificação falhou" }, { status: 403 });
}

// POST — Receber notificações (mensagens, status, etc.)
export async function POST(req: NextRequest) {
	// Validação HMAC (opcional mas recomendado)
	if (APP_SECRET) {
		const signature = req.headers.get("x-hub-signature-256");
		const rawBody = await req.clone().text();

		if (signature) {
			const expectedSig =
				"sha256=" +
				crypto.createHmac("sha256", APP_SECRET).update(rawBody).digest("hex");

			if (signature !== expectedSig) {
				console.warn("❌ Assinatura HMAC inválida");
				return NextResponse.json(
					{ error: "Invalid signature" },
					{ status: 401 },
				);
			}
		}
	}

	const body = await req.json();

	// Processar cada entry
	const entries = body.entry ?? [];

	for (const entry of entries) {
		const changes = entry.changes ?? [];

		for (const change of changes) {
			const value = change.value;

			// Mensagens recebidas
			if (value.messages) {
				for (const message of value.messages) {
					const from = message.from; // número do remetente
					const type = message.type;

					console.log(`📩 Mensagem recebida de ${from} (tipo: ${type})`);

					if (type === "text") {
						console.log(`   Texto: ${message.text.body}`);
					}

					// Aqui você processa a mensagem:
					// - Salvar no banco
					// - Responder automaticamente
					// - Disparar notificação
				}
			}

			// Status de mensagens enviadas (sent, delivered, read)
			if (value.statuses) {
				for (const status of value.statuses) {
					console.log(
						`📊 Status: ${status.status} para ${status.recipient_id} (msg: ${status.id})`,
					);
				}
			}
		}
	}

	// IMPORTANTE: sempre retornar 200, senão a Meta reenvia
	return NextResponse.json({ status: "ok" }, { status: 200 });
}
