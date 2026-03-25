"use client";

import { Download, FileSpreadsheet, FileText, Loader2 } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useMonthContext } from "@/components/layout/month-context";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { formatCurrency, getCurrentMonth } from "@/lib/utils";

type Scope = "month" | "year" | "all";
type Format = "xlsx" | "ods" | "pdf";

interface ExportRow {
	id: number;
	date: string;
	description: string;
	amount: number;
	type: string;
	categoryName: string | null;
	notes: string | null;
}

interface ExportResult {
	data: ExportRow[];
	totals: {
		income: number;
		expense: number;
		balance: number;
		count: number;
	};
}

const MONTHS = [
	{ value: "1", label: "Janeiro" },
	{ value: "2", label: "Fevereiro" },
	{ value: "3", label: "Março" },
	{ value: "4", label: "Abril" },
	{ value: "5", label: "Maio" },
	{ value: "6", label: "Junho" },
	{ value: "7", label: "Julho" },
	{ value: "8", label: "Agosto" },
	{ value: "9", label: "Setembro" },
	{ value: "10", label: "Outubro" },
	{ value: "11", label: "Novembro" },
	{ value: "12", label: "Dezembro" },
];

const currentYear = getCurrentMonth().year;
const YEARS = Array.from({ length: currentYear - 2019 }, (_, i) =>
	String(currentYear - i),
);

function buildPeriodLabel(scope: Scope, month: string, year: string): string {
	if (scope === "month") {
		const monthLabel = MONTHS.find((m) => m.value === month)?.label ?? month;
		return `${monthLabel} de ${year}`;
	}
	if (scope === "year") return `Ano ${year}`;
	return "Todos os lançamentos";
}

async function exportXlsxOds(
	rows: ExportRow[],
	totals: ExportResult["totals"],
	format: "xlsx" | "ods",
	filename: string,
) {
	const XLSX = await import("xlsx");

	const sheetData = [
		["Data", "Descrição", "Tipo", "Categoria", "Valor (R$)", "Observações"],
		...rows.map((r) => [
			r.date,
			r.description,
			r.type === "income" ? "Receita" : "Despesa",
			r.categoryName ?? "",
			r.type === "expense" ? -r.amount : r.amount,
			r.notes ?? "",
		]),
		[],
		["", "", "", "Receitas", totals.income, ""],
		["", "", "", "Despesas", totals.expense, ""],
		["", "", "", "Saldo", totals.balance, ""],
	];

	const wb = XLSX.utils.book_new();
	const ws = XLSX.utils.aoa_to_sheet(sheetData);

	// Column widths
	ws["!cols"] = [
		{ wch: 12 },
		{ wch: 40 },
		{ wch: 10 },
		{ wch: 20 },
		{ wch: 14 },
		{ wch: 30 },
	];

	XLSX.utils.book_append_sheet(wb, ws, "Lançamentos");
	XLSX.writeFile(wb, `${filename}.${format}`);
}

async function exportPdf(
	rows: ExportRow[],
	totals: ExportResult["totals"],
	periodLabel: string,
	filename: string,
) {
	const { jsPDF } = await import("jspdf");
	const autoTable = (await import("jspdf-autotable")).default;

	const doc = new jsPDF({ orientation: "landscape" });

	// Title
	doc.setFontSize(16);
	doc.setTextColor(40, 40, 40);
	doc.text("Lançamentos — ReConta", 14, 16);

	doc.setFontSize(10);
	doc.setTextColor(100, 100, 100);
	doc.text(`Período: ${periodLabel}`, 14, 23);
	doc.text(
		`Gerado em: ${new Intl.DateTimeFormat("pt-BR").format(new Date())}`,
		14,
		28,
	);

	// Summary box
	autoTable(doc, {
		startY: 33,
		head: [["Receitas", "Despesas", "Saldo", "Total de registros"]],
		body: [
			[
				formatCurrency(totals.income),
				formatCurrency(totals.expense),
				formatCurrency(totals.balance),
				String(totals.count),
			],
		],
		theme: "grid",
		headStyles: { fillColor: [79, 70, 229], fontSize: 9 },
		bodyStyles: { fontSize: 9 },
		columnStyles: {
			2: {
				textColor: totals.balance >= 0 ? [22, 163, 74] : [220, 38, 38],
			},
		},
	});

	const summaryEnd = (doc as unknown as { lastAutoTable: { finalY: number } })
		.lastAutoTable.finalY;

	// Transactions table
	autoTable(doc, {
		startY: summaryEnd + 8,
		head: [["Data", "Descrição", "Tipo", "Categoria", "Valor", "Obs."]],
		body: rows.map((r) => [
			r.date,
			r.description,
			r.type === "income" ? "Receita" : "Despesa",
			r.categoryName ?? "—",
			(r.type === "expense" ? "- " : "+ ") + formatCurrency(r.amount),
			r.notes ?? "",
		]),
		theme: "striped",
		headStyles: { fillColor: [39, 39, 42], fontSize: 8 },
		bodyStyles: { fontSize: 7.5 },
		columnStyles: {
			0: { cellWidth: 22 },
			1: { cellWidth: 80 },
			2: { cellWidth: 20 },
			3: { cellWidth: 35 },
			4: { cellWidth: 28, halign: "right" },
			5: { cellWidth: 50 },
		},
		didParseCell(data) {
			if (data.column.index === 4 && data.section === "body") {
				const text = String(data.cell.raw);
				data.cell.styles.textColor = text.startsWith("-")
					? [220, 38, 38]
					: [22, 163, 74];
			}
		},
	});

	doc.save(`${filename}.pdf`);
}

export function ExportarClient() {
	const { month: curMonth, year: curYear } = useMonthContext();

	const [scope, setScope] = useState<Scope>("month");
	const [month, setMonth] = useState(String(curMonth));
	const [year, setYear] = useState(String(curYear));
	const [format, setFormat] = useState<Format>("xlsx");
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	async function handleExport() {
		setLoading(true);
		setError(null);

		try {
			const params = new URLSearchParams({ scope });
			if (scope === "month") {
				params.set("month", month);
				params.set("year", year);
			} else if (scope === "year") {
				params.set("year", year);
			}

			const res = await fetch(`/api/export?${params}`);
			if (!res.ok) throw new Error("Erro ao buscar dados para exportação");

			const result: ExportResult = await res.json();

			if (result.totals.count === 0) {
				setError("Nenhum lançamento encontrado para o período selecionado.");
				return;
			}

			const periodLabel = buildPeriodLabel(scope, month, year);
			const safePeriod = periodLabel
				.toLowerCase()
				.replace(/\s+/g, "-")
				.replace(/[^a-z0-9-]/g, "");
			const filename = `lancamentos-${safePeriod}`;

			if (format === "pdf") {
				await exportPdf(result.data, result.totals, periodLabel, filename);
			} else {
				await exportXlsxOds(result.data, result.totals, format, filename);
			}
		} catch (err) {
			setError(err instanceof Error ? err.message : "Erro inesperado");
		} finally {
			setLoading(false);
		}
	}

	return (
		<div className="max-w-xl space-y-8">
			{/* Period */}
			<section className="rounded-xl border border-zinc-800 bg-zinc-900 p-6 space-y-4">
				<h2 className="text-sm font-semibold text-zinc-300 uppercase tracking-wide">
					Período
				</h2>

				<div className="space-y-1">
					<label htmlFor="scope" className="text-xs text-zinc-400">
						Filtrar por
					</label>
					<Select value={scope} onValueChange={(v) => setScope(v as Scope)}>
						<SelectTrigger id="scope">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="month">Mês</SelectItem>
							<SelectItem value="year">Ano</SelectItem>
							<SelectItem value="all">Todos</SelectItem>
						</SelectContent>
					</Select>
				</div>

				{scope === "month" && (
					<div className="grid grid-cols-2 gap-3">
						<div className="space-y-1">
							<label htmlFor="month" className="text-xs text-zinc-400">
								Mês
							</label>
							<Select value={month} onValueChange={setMonth}>
								<SelectTrigger id="month">
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									{MONTHS.map((m) => (
										<SelectItem key={m.value} value={m.value}>
											{m.label}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>
						<div className="space-y-1">
							<label htmlFor="year" className="text-xs text-zinc-400">
								Ano
							</label>
							<Select value={year} onValueChange={setYear}>
								<SelectTrigger id="year">
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									{YEARS.map((y) => (
										<SelectItem key={y} value={y}>
											{y}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>
					</div>
				)}

				{scope === "year" && (
					<div className="space-y-1">
						<label htmlFor="year" className="text-xs text-zinc-400">
							Ano
						</label>
						<Select value={year} onValueChange={setYear}>
							<SelectTrigger id="year">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								{YEARS.map((y) => (
									<SelectItem key={y} value={y}>
										{y}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>
				)}
			</section>

			{/* Format */}
			<section className="rounded-xl border border-zinc-800 bg-zinc-900 p-6 space-y-4">
				<h2 className="text-sm font-semibold text-zinc-300 uppercase tracking-wide">
					Formato
				</h2>

				<div className="grid grid-cols-3 gap-3">
					{(["xlsx", "ods", "pdf"] as Format[]).map((f) => {
						const Icon = f === "pdf" ? FileText : FileSpreadsheet;
						const labels: Record<Format, string> = {
							xlsx: "Excel (.xlsx)",
							ods: "Calc (.ods)",
							pdf: "PDF (.pdf)",
						};
						return (
							<button
								key={f}
								type="button"
								onClick={() => setFormat(f)}
								className={`flex flex-col items-center gap-2 rounded-lg border p-4 text-sm font-medium transition-colors ${
									format === f
										? "border-indigo-500 bg-indigo-600/20 text-indigo-300"
										: "border-zinc-700 bg-zinc-800 text-zinc-400 hover:border-zinc-500 hover:text-zinc-200"
								}`}
							>
								<Icon className="h-6 w-6" />
								{labels[f]}
							</button>
						);
					})}
				</div>
			</section>

			{error && (
				<p className="rounded-lg border border-red-800 bg-red-950/50 px-4 py-3 text-sm text-red-400">
					{error}
				</p>
			)}

			<Button
				onClick={handleExport}
				disabled={loading}
				className="w-full"
				size="lg"
			>
				{loading ? (
					<>
						<Loader2 className="mr-2 h-4 w-4 animate-spin" />
						Exportando…
					</>
				) : (
					<>
						<Download className="mr-2 h-4 w-4" />
						Exportar lançamentos
					</>
				)}
			</Button>
		</div>
	);
}
