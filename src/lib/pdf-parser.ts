"use server";

export interface ParsedTransaction {
	date: string;
	description: string;
	amount: number;
	type: "income" | "expense";
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
): Promise<ParsedTransaction[]> {
	const pdfjs = await import("pdfjs-dist/legacy/build/pdf.mjs");
	// In Node environment, PDF.js uses a fake worker and tries to import pdf.worker.mjs
	// relatively from the calling module, which breaks in Next.js chunks.
	// Provide the worker handler directly via globalThis.pdfjsWorker to avoid dynamic import.
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
	try {
		for (let i = 1; i <= pdf.numPages; i++) {
			const page = await pdf.getPage(i);
			const content = await page.getTextContent();
			const pageText = content.items
				.map((item) => ("str" in item ? (item.str ?? "") : ""))
				.join(" ");
			page.cleanup();
			parseTextToTransactions(pageText, transactions);
		}
	} finally {
		pdf.destroy();
	}

	return transactions;
}

function parseTextToTransactions(
	text: string,
	transactions: ParsedTransaction[] = [],
): ParsedTransaction[] {
	// Pattern: DD/MM/YYYY ... description ... value
	// Tries multiple common Brazilian bank statement formats
	const patterns = [
		// Itaú / Bradesco / BB style: DD/MM/YYYY description value
		/(\d{2}\/\d{2}\/\d{4})\s+(.+?)\s+([-]?\d{1,3}(?:\.\d{3})*(?:,\d{2})?)\s*([DC]?)/gi,
		// Generic: date description +/-amount
		/(\d{2}\/\d{2}\/\d{4})\s+(.+?)\s+([+-]?\d+[.,]\d{2})/gi,
	];

	for (const pattern of patterns) {
		pattern.lastIndex = 0;
		let match = pattern.exec(text);
		while (match !== null) {
			const dateStr = match[1];
			const description = match[2].trim().replace(/\s+/g, " ");
			const rawAmount = match[3].replace(/\./g, "").replace(",", ".");
			const debitCreditFlag = match[4] || "";
			match = pattern.exec(text);

			const amount = Math.abs(Number.parseFloat(rawAmount));
			if (Number.isNaN(amount) || amount === 0) continue;

			const [day, month, year] = dateStr.split("/");
			const isoDate = `${year}-${month}-${day}`;

			let type: "income" | "expense";
			if (debitCreditFlag.toUpperCase() === "D") {
				type = "expense";
			} else if (debitCreditFlag.toUpperCase() === "C") {
				type = "income";
			} else if (Number.parseFloat(rawAmount) < 0) {
				type = "expense";
			} else {
				type = "income";
			}

			const isSaldoLine = /\bsaldo\b/i.test(description);
			if (!isSaldoLine && description.length > 3 && description.length < 200) {
				transactions.push({ date: isoDate, description, amount, type });
			}
		}
		if (transactions.length > 0) break;
	}

	return transactions;
}
