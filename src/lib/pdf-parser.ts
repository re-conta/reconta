"use server";

export interface ParsedTransaction {
	date: string;
	description: string;
	amount: number;
	type: "income" | "expense";
}

export async function parseBankStatementPdf(
	buffer: ArrayBuffer,
): Promise<ParsedTransaction[]> {
	const { getDocument } = await import("pdfjs-dist/legacy/build/pdf.mjs");

	const loadingTask = getDocument({ data: buffer });
	const pdf = await loadingTask.promise;

	let fullText = "";
	for (let i = 1; i <= pdf.numPages; i++) {
		const page = await pdf.getPage(i);
		const content = await page.getTextContent();
		const pageText = content.items
			.map((item) => ("str" in item ? (item.str ?? "") : ""))
			.join(" ");
		fullText += `${pageText}\n`;
	}

	return parseTextToTransactions(fullText);
}

function parseTextToTransactions(text: string): ParsedTransaction[] {
	const transactions: ParsedTransaction[] = [];

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
		const match = pattern.exec(text);
		while (match !== null) {
			const dateStr = match[1];
			const description = match[2].trim().replace(/\s+/g, " ");
			const rawAmount = match[3].replace(/\./g, "").replace(",", ".");
			const debitCreditFlag = match[4] || "";

			const amount = Math.abs(Number.parseFloat(rawAmount));
			if (Number.isNaN(amount) || amount === 0) continue;

			const [day, month, year] = dateStr.split("/");
			const isoDate = `${year}-${month}-${day}`;

			let type: "income" | "expense";
			if (
				debitCreditFlag.toUpperCase() === "C" ||
				Number.parseFloat(rawAmount) > 0
			) {
				type = "income";
			} else {
				type = "expense";
			}

			if (description.length > 3 && description.length < 200) {
				transactions.push({ date: isoDate, description, amount, type });
			}
		}
		if (transactions.length > 0) break;
	}

	return transactions;
}
