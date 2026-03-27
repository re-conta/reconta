"use server";

export interface ParsedTransaction {
	date: string;
	description: string;
	amount: number;
	type: "income" | "expense";
	pixBeneficiary: string | null;
}

export interface ParsedStatement {
	bank: string | null;
	transactions: ParsedTransaction[];
}

declare global {
	interface GlobalThis {
		pdfjsWorker?: {
			WorkerMessageHandler: unknown;
		};
	}
}

export async function parseBankStatementPdf(
	buffer: ArrayBuffer,
): Promise<ParsedStatement> {
	const pdfjs = await import("pdfjs-dist/legacy/build/pdf.mjs");
	type PdfjsWorkerType = { WorkerMessageHandler: unknown };
	const globalThisPdfjs = globalThis as typeof globalThis & {
		pdfjsWorker?: PdfjsWorkerType;
	};

	if (!globalThisPdfjs.pdfjsWorker?.WorkerMessageHandler) {
		const worker = await import("pdfjs-dist/legacy/build/pdf.worker.mjs");
		globalThisPdfjs.pdfjsWorker = {
			WorkerMessageHandler: worker.WorkerMessageHandler,
		};
	}

	type PdfjsDocumentInit = {
		data: ArrayBuffer;
		standardFontDataUrl: string;
	};

	const standardFontDataUrl = `file://${process.cwd()}/node_modules/pdfjs-dist/standard_fonts/`;
	const loadingTask = pdfjs.getDocument({
		data: buffer,
		standardFontDataUrl,
	} as unknown as PdfjsDocumentInit);
	const pdf = await loadingTask.promise;

	const transactions: ParsedTransaction[] = [];
	const allText: string[] = [];
	try {
		for (let i = 1; i <= pdf.numPages; i++) {
			const page = await pdf.getPage(i);
			const content = await page.getTextContent();
			const pageText = content.items
				.map((item) => ("str" in item ? (item.str ?? "") : ""))
				.join(" ");
			page.cleanup();
			allText.push(pageText);
			parseTextToTransactions(pageText, transactions);
		}
	} finally {
		pdf.destroy();
	}

	const fullText = allText.join(" ");
	const bank = detectBank(fullText);

	return { bank, transactions };
}

/** Detect which bank issued the statement based on full PDF text */
const BANK_PATTERNS: Array<{ name: string; pattern: RegExp }> = [
	{
		name: "Itaú",
		pattern: /\b(ita[uú]\s*unibanco|banco\s*ita[uú]|itau\.com\.br)\b/i,
	},
	{ name: "Bradesco", pattern: /\b(bradesco|banco\s*bradesco)\b/i },
	{
		name: "Banco do Brasil",
		pattern: /\b(banco\s*do\s*brasil|bb\.com\.br)\b/i,
	},
	{
		name: "Caixa",
		pattern: /\b(caixa\s*econ[oô]mica|caixa\s*federal|cef\b|caixa\.gov\.br)\b/i,
	},
	{ name: "Nubank", pattern: /\b(nubank|nu\s*pagamentos|nubank\.com\.br)\b/i },
	{ name: "Santander", pattern: /\b(santander|banco\s*santander)\b/i },
	{ name: "Inter", pattern: /\b(banco\s*inter|inter\.co|bancointer)\b/i },
	{ name: "C6 Bank", pattern: /\b(c6\s*bank|c6bank|banco\s*c6)\b/i },
	{ name: "Sicoob", pattern: /\b(sicoob|bancoob)\b/i },
	{ name: "Sicredi", pattern: /\b(sicredi)\b/i },
	{ name: "BTG Pactual", pattern: /\b(btg\s*pactual)\b/i },
	{ name: "Safra", pattern: /\b(banco\s*safra|safra\.com\.br)\b/i },
	{ name: "Original", pattern: /\b(banco\s*original|original\.com\.br)\b/i },
	{ name: "PagBank", pattern: /\b(pagbank|pagseguro)\b/i },
	{ name: "Mercado Pago", pattern: /\b(mercado\s*pago)\b/i },
	{ name: "Neon", pattern: /\b(banco\s*neon|neon\.com\.br)\b/i },
	{ name: "Next", pattern: /\b(banco\s*next|next\.me)\b/i },
	{ name: "Banrisul", pattern: /\b(banrisul)\b/i },
	{ name: "BRB", pattern: /\b(brb\b|banco\s*de\s*bras[ií]lia)\b/i },
	{ name: "Daycoval", pattern: /\b(daycoval)\b/i },
];

function detectBank(text: string): string | null {
	for (const { name, pattern } of BANK_PATTERNS) {
		if (pattern.test(text)) return name;
	}
	return null;
}

/**
 * Extract PIX beneficiary from a transaction description.
 * Common patterns:
 * - "Transferencia PIX - 12345 NOME DO DESTINATARIO"
 * - "PIX enviado - NOME DO DESTINATARIO"
 * - "PIX recebido - NOME DO REMETENTE"
 * - "Pix Enviado NOME"
 */
const PIX_BENEFICIARY_PATTERNS = [
	/transfer[eê]ncia\s+pix\s*[-–]\s*\d+\s+(.+)/i,
	/transfer[eê]ncia\s+pix\s*[-–]\s*(.+)/i,
	/pix\s+enviad[oa]\s*[-–]\s*(.+)/i,
	/pix\s+recebid[oa]\s*[-–]\s*(.+)/i,
	/pix\s+enviad[oa]\s+(.+)/i,
	/pix\s+recebid[oa]\s+(.+)/i,
];

function extractPixBeneficiary(description: string): string | null {
	for (const pattern of PIX_BENEFICIARY_PATTERNS) {
		const match = pattern.exec(description);
		if (match?.[1]) {
			const name = match[1].trim().replace(/\s+/g, " ");
			if (name.length >= 2 && name.length < 150) return name;
		}
	}
	return null;
}

/** Keywords that indicate an income transaction */
const INCOME_KEYWORDS =
	/\b(credito|crédito|deposito|depósito|transferencia\s+recebida|pix\s+recebid[oa]|ted\s+recebid[oa]|doc\s+recebid[oa]|estorno|reembolso|rendimento|dividendo|salario|salário|cashback|devolucao|devolução|resgate|bonificacao|bonificação)\b/i;

/** Keywords that indicate an expense transaction */
const EXPENSE_KEYWORDS =
	/\b(debito|débito|pagamento|pag\b|pix\s+enviad[oa]|transferencia\s+enviad[oa]|ted\s+enviad[oa]|doc\s+enviad[oa]|compra|saque|tarifa|taxa|anuidade|iof|juros|multa|encargo|seguro|mensalidade|assinatura|boleto|fatura|parcela)\b/i;

function inferTypeFromDescription(description: string): "income" | "expense" {
	if (INCOME_KEYWORDS.test(description)) return "income";
	if (EXPENSE_KEYWORDS.test(description)) return "expense";
	// Default to expense — most bank statement lines are debits
	return "expense";
}

function parseTextToTransactions(
	text: string,
	transactions: ParsedTransaction[] = [],
): ParsedTransaction[] {
	// Pattern: DD/MM/YYYY ... description ... value
	// Tries multiple common Brazilian bank statement formats
	const patterns = [
		// Itaú / Bradesco / BB style: DD/MM/YYYY description value [D/C]
		/(\d{2}\/\d{2}\/\d{4})\s+(.+?)\s+([+-]?\s*\d{1,3}(?:\.\d{3})*(?:,\d{2})?)\s*([DC]?)/gi,
		// Generic: date description +/-amount
		/(\d{2}\/\d{2}\/\d{4})\s+(.+?)\s+([+-]?\s*\d+[.,]\d{2})/gi,
		// Amount in parentheses (negative): DD/MM/YYYY description (amount)
		/(\d{2}\/\d{2}\/\d{4})\s+(.+?)\s+\((\d{1,3}(?:\.\d{3})*(?:,\d{2})?)\)/gi,
	];

	for (const pattern of patterns) {
		pattern.lastIndex = 0;
		let match = pattern.exec(text);
		const isParenthesesPattern = pattern.source.includes("\\(");
		while (match !== null) {
			const dateStr = match[1];
			const description = match[2].trim().replace(/\s+/g, " ");
			const rawAmount = match[3]
				.replace(/\s/g, "")
				.replace(/\./g, "")
				.replace(",", ".");
			const debitCreditFlag = match[4] || "";
			match = pattern.exec(text);

			const numericAmount = Number.parseFloat(rawAmount);
			const amount = Math.abs(numericAmount);
			if (Number.isNaN(amount) || amount === 0) continue;

			const [day, month, year] = dateStr.split("/");
			const isoDate = `${year}-${month}-${day}`;

			let type: "income" | "expense";
			if (isParenthesesPattern) {
				// Amounts in parentheses are always negative/expense
				type = "expense";
			} else if (debitCreditFlag.toUpperCase() === "D") {
				type = "expense";
			} else if (debitCreditFlag.toUpperCase() === "C") {
				type = "income";
			} else if (numericAmount < 0) {
				type = "expense";
			} else if (rawAmount.startsWith("+")) {
				type = "income";
			} else {
				// No explicit sign or flag — infer from description keywords
				type = inferTypeFromDescription(description);
			}

			const isSaldoLine = /\bsaldo\b/i.test(description);
			if (!isSaldoLine && description.length > 3 && description.length < 200) {
				const pixBeneficiary = extractPixBeneficiary(description);
				transactions.push({
					date: isoDate,
					description,
					amount,
					type,
					pixBeneficiary,
				});
			}
		}
		if (transactions.length > 0) break;
	}

	return transactions;
}
