const API_VERSION = process.env.WHATSAPP_API_VERSION || "v21.0";
const PHONE_NUMBER_ID = process.env.WHATSAPP_PHONE_NUMBER_ID!;
const ACCESS_TOKEN = process.env.WHATSAPP_TOKEN!;

const BASE_URL = `https://graph.facebook.com/${API_VERSION}/${PHONE_NUMBER_ID}`;

interface SendMessageResponse {
	messaging_product: string;
	contacts: { input: string; wa_id: string }[];
	messages: { id: string }[];
}

export async function sendTextMessage(
	to: string,
	body: string,
): Promise<SendMessageResponse> {
	const res = await fetch(`${BASE_URL}/messages`, {
		method: "POST",
		headers: {
			Authorization: `Bearer ${ACCESS_TOKEN}`,
			"Content-Type": "application/json",
		},
		body: JSON.stringify({
			messaging_product: "whatsapp",
			recipient_type: "individual",
			to,
			type: "text",
			text: { preview_url: false, body },
		}),
	});

	if (!res.ok) {
		const error = await res.json();
		throw new Error(`WhatsApp API error: ${JSON.stringify(error)}`);
	}

	return res.json();
}

export async function sendTemplateMessage(
	to: string,
	templateName: string = "hello_world",
	languageCode: string = "en_US",
): Promise<SendMessageResponse> {
	const res = await fetch(`${BASE_URL}/messages`, {
		method: "POST",
		headers: {
			Authorization: `Bearer ${ACCESS_TOKEN}`,
			"Content-Type": "application/json",
		},
		body: JSON.stringify({
			messaging_product: "whatsapp",
			to,
			type: "template",
			template: {
				name: templateName,
				language: { code: languageCode },
			},
		}),
	});

	if (!res.ok) {
		const error = await res.json();
		throw new Error(`WhatsApp API error: ${JSON.stringify(error)}`);
	}

	return res.json();
}
