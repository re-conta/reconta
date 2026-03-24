export const dynamic = "force-dynamic";

import { NextResponse } from "next/server";
import { db } from "@/lib/db";
import { pdfImports, transactions } from "@/lib/db/schema";
import { requireSession } from "@/lib/auth-session";
import { parseBankStatementPdf } from "@/lib/pdf-parser";

export async function POST(request: Request) {
	try {
		const { userId, unauthorized } = await requireSession();
		if (unauthorized) return unauthorized;

		const formData = await request.formData();
		const file = formData.get("file") as File | null;
		const accountId = formData.get("accountId") as string | null;
		const defaultCategoryId = formData.get("categoryId") as string | null;

		if (!file) {
			return NextResponse.json({ error: "Arquivo não enviado" }, { status: 400 });
		}

		if (!file.name.endsWith(".pdf")) {
			return NextResponse.json(
				{ error: "Apenas arquivos PDF são suportados" },
				{ status: 400 },
			);
		}

		const buffer = await file.arrayBuffer();
		const parsed = await parseBankStatementPdf(buffer);

		if (parsed.length === 0) {
			return NextResponse.json(
				{
					error:
						"Nenhuma transação encontrada no PDF. Verifique se o extrato está no formato correto.",
				},
				{ status: 422 },
			);
		}

		const inserted = await db
			.insert(transactions)
			.values(
				parsed.map((t) => ({
					userId,
					date: t.date,
					description: t.description,
					amount: t.amount,
					type: t.type,
					categoryId: defaultCategoryId ? Number(defaultCategoryId) : null,
					accountId: accountId ? Number(accountId) : null,
					importedFrom: file.name,
				})),
			)
			.returning();

		await db.insert(pdfImports).values({
			userId,
			filename: file.name,
			accountId: accountId ? Number(accountId) : null,
			transactionCount: inserted.length,
		});

		return NextResponse.json({
			imported: inserted.length,
			transactions: inserted,
		});
	} catch (error) {
		console.error("Erro ao importar PDF:", error);
		return NextResponse.json(
			{ error: "Erro interno ao processar o arquivo. Tente novamente." },
			{ status: 500 },
		);
	}
}
