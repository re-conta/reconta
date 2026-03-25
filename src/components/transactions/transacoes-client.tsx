"use client";

import {
	ChevronDown,
	ChevronLeft,
	ChevronRight,
	Pencil,
	Plus,
	Search,
	Trash2,
} from "lucide-react";
import { useEffect, useRef, useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { useMonthContext } from "@/components/layout/month-context";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { formatCurrency, formatDate, formatMonth } from "@/lib/utils";
import { TransactionDialog } from "./transaction-dialog";
import { BulkEditDialog, type BulkEditFields } from "./bulk-edit-dialog";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface Transaction {
	id: number;
	date: string;
	description: string;
	amount: number;
	type: string;
	categoryId: number | null;
	categoryName: string | null;
	categoryColor: string | null;
	notes: string | null;
	accountId: number | null;
}

interface Totals {
	income: number;
	expense: number;
	balance: number;
	count: number;
}

export function TransacoesClient() {
	const { month, year, setPeriod } = useMonthContext();
	const [search, setSearch] = useState("");
	const [typeFilter, setTypeFilter] = useState<"all" | "income" | "expense">(
		"all",
	);
	const [transactions, setTransactions] = useState<Transaction[]>([]);
	const [totals, setTotals] = useState<Totals>({
		income: 0,
		expense: 0,
		balance: 0,
		count: 0,
	});
	const [loading, setLoading] = useState(true);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [editingTx, setEditingTx] = useState<Transaction | null>(null);
	const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
	const [bulkEditOpen, setBulkEditOpen] = useState(false);

	const selectAllRef = useRef<HTMLInputElement>(null);

	const today = new Date();
	const isCurrentMonth =
		month === today.getMonth() + 1 && year === today.getFullYear();

	// Indeterminate state for select-all checkbox
	useEffect(() => {
		if (selectAllRef.current) {
			const some = selectedIds.size > 0;
			const all =
				transactions.length > 0 && selectedIds.size === transactions.length;
			selectAllRef.current.indeterminate = some && !all;
		}
	}, [selectedIds, transactions]);

	const fetchTransactions = useCallback(() => {
		setLoading(true);
		const params = new URLSearchParams({
			month: String(month),
			year: String(year),
			limit: "200",
		});
		if (typeFilter !== "all") params.set("type", typeFilter);
		if (search) params.set("search", search);

		fetch(`/api/transactions?${params}`)
			.then((r) => r.json())
			.then((d) => {
				setTransactions(d.data ?? []);
				setTotals(d.totals ?? { income: 0, expense: 0, balance: 0, count: 0 });
				setLoading(false);
			});
	}, [month, year, typeFilter, search]);

	useEffect(() => {
		fetchTransactions();
	}, [fetchTransactions]);

	// Debounce search
	useEffect(() => {
		const t = setTimeout(() => fetchTransactions(), 300);
		return () => clearTimeout(t);
	}, [fetchTransactions]);

	function prevMonth() {
		setSelectedIds(new Set());
		if (month === 1) setPeriod(12, year - 1);
		else setPeriod(month - 1, year);
	}

	function nextMonth() {
		setSelectedIds(new Set());
		if (month === 12) setPeriod(1, year + 1);
		else setPeriod(month + 1, year);
	}

	function toggleAll() {
		if (selectedIds.size === transactions.length) {
			setSelectedIds(new Set());
		} else {
			setSelectedIds(new Set(transactions.map((t) => t.id)));
		}
	}

	function toggleId(id: number) {
		setSelectedIds((prev) => {
			const next = new Set(prev);
			if (next.has(id)) next.delete(id);
			else next.add(id);
			return next;
		});
	}

	async function deleteTransaction(id: number) {
		if (!confirm("Deseja excluir este lançamento?")) return;
		await fetch(`/api/transactions/${id}`, { method: "DELETE" });
		fetchTransactions();
	}

	async function bulkDelete(scope: "month" | "year" | "all") {
		const labels: Record<typeof scope, string> = {
			month: formatMonth(month, year),
			year: String(year),
			all: "todos os lançamentos",
		};
		if (
			!confirm(
				`Deseja excluir ${labels[scope]}? Esta ação não pode ser desfeita.`,
			)
		)
			return;
		await fetch("/api/transactions", {
			method: "DELETE",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ scope, month, year }),
		});
		fetchTransactions();
	}

	async function bulkDeleteSelected() {
		if (
			!confirm(
				`Deseja excluir ${selectedIds.size} lançamento${selectedIds.size !== 1 ? "s" : ""}? Esta ação não pode ser desfeita.`,
			)
		)
			return;
		await Promise.all(
			Array.from(selectedIds).map((id) =>
				fetch(`/api/transactions/${id}`, { method: "DELETE" }),
			),
		);
		setSelectedIds(new Set());
		fetchTransactions();
	}

	async function bulkEditSelected(fields: BulkEditFields) {
		await fetch("/api/transactions", {
			method: "PATCH",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ ids: Array.from(selectedIds), fields }),
		});
		setSelectedIds(new Set());
		fetchTransactions();
	}

	return (
		<div className="space-y-4">
			{/* Controls */}
			<div className="flex flex-wrap items-center gap-3">
				<div className="flex items-center gap-2">
					<Button variant="outline" size="icon" onClick={prevMonth}>
						<ChevronLeft className="h-4 w-4" />
					</Button>
					<span className="font-medium capitalize min-w-40 text-center text-zinc-100">
						{formatMonth(month, year)}
					</span>
					<Button
						variant="outline"
						size="icon"
						onClick={nextMonth}
						disabled={isCurrentMonth}
					>
						<ChevronRight className="h-4 w-4" />
					</Button>
				</div>

				<div className="relative flex-1 min-w-50">
					<Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
					<Input
						placeholder="Buscar..."
						value={search}
						onChange={(e) => setSearch(e.target.value)}
						className="pl-9"
					/>
				</div>

				<div className="flex gap-1">
					{(["all", "income", "expense"] as const).map((t) => (
						<Button
							key={t}
							variant={typeFilter === t ? "default" : "outline"}
							size="sm"
							onClick={() => setTypeFilter(t)}
						>
							{t === "all" ? "Todos" : t === "income" ? "Receitas" : "Despesas"}
						</Button>
					))}
				</div>

				<div className="ml-auto flex gap-2">
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button variant="outline" size="sm">
								<Trash2 className="h-4 w-4 text-red-400" />
								Apagar em massa
								<ChevronDown className="h-3.5 w-3.5" />
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="end">
							<DropdownMenuItem
								className="text-red-400 focus:text-red-300"
								onSelect={() => bulkDelete("month")}
							>
								Apagar {formatMonth(month, year)}
							</DropdownMenuItem>
							<DropdownMenuItem
								className="text-red-400 focus:text-red-300"
								onSelect={() => bulkDelete("year")}
							>
								Apagar ano {year}
							</DropdownMenuItem>
							<DropdownMenuSeparator />
							<DropdownMenuItem
								className="text-red-400 focus:text-red-300"
								onSelect={() => bulkDelete("all")}
							>
								Apagar tudo
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>

					<Button
						onClick={() => {
							setEditingTx(null);
							setDialogOpen(true);
						}}
					>
						<Plus className="h-4 w-4" />
						Novo lançamento
					</Button>
				</div>
			</div>

			{/* Totals */}
			<div className="grid grid-cols-3 gap-3">
				<div className="rounded-lg bg-emerald-900/20 border border-emerald-800/30 p-4">
					<p className="text-xs text-zinc-400 mb-1">Receitas</p>
					<p className="text-lg font-bold text-emerald-400">
						{formatCurrency(totals.income)}
					</p>
				</div>
				<div className="rounded-lg bg-red-900/20 border border-red-800/30 p-4">
					<p className="text-xs text-zinc-400 mb-1">Despesas</p>
					<p className="text-lg font-bold text-red-400">
						{formatCurrency(totals.expense)}
					</p>
				</div>
				<div
					className={`rounded-lg border p-4 ${totals.balance >= 0 ? "bg-indigo-900/20 border-indigo-800/30" : "bg-red-900/20 border-red-800/30"}`}
				>
					<p className="text-xs text-zinc-400 mb-1">Saldo</p>
					<p
						className={`text-lg font-bold ${totals.balance >= 0 ? "text-indigo-400" : "text-red-400"}`}
					>
						{formatCurrency(totals.balance)}
					</p>
				</div>
			</div>

			{/* Bulk action bar */}
			{selectedIds.size > 0 && (
				<div className="flex flex-wrap items-center gap-3 px-4 py-2.5 rounded-lg bg-indigo-900/30 border border-indigo-700/40 text-sm">
					<span className="text-indigo-300 font-medium">
						{selectedIds.size} selecionado{selectedIds.size !== 1 ? "s" : ""}
					</span>
					<div className="flex flex-wrap gap-2 ml-auto">
						<Button
							size="sm"
							variant="outline"
							onClick={() => setBulkEditOpen(true)}
						>
							<Pencil className="h-3.5 w-3.5" />
							Editar selecionados
						</Button>
						<Button
							size="sm"
							variant="outline"
							className="text-red-400 hover:text-red-300 border-red-800/50"
							onClick={bulkDeleteSelected}
						>
							<Trash2 className="h-3.5 w-3.5" />
							Excluir selecionados
						</Button>
						<Button
							size="sm"
							variant="ghost"
							onClick={() => setSelectedIds(new Set())}
						>
							Cancelar
						</Button>
					</div>
				</div>
			)}

			{/* Table */}
			<Card>
				<CardContent className="p-0">
					{loading ? (
						<div className="p-8 text-center text-zinc-500 text-sm">
							Carregando...
						</div>
					) : transactions.length === 0 ? (
						<div className="p-8 text-center text-zinc-500 text-sm">
							Nenhum lançamento encontrado.
						</div>
					) : (
						<div className="overflow-x-auto">
							<table className="w-full text-sm">
								<thead>
									<tr className="border-b border-zinc-800">
										<th className="px-4 py-3 w-10">
											<Checkbox
												ref={selectAllRef}
												checked={
													transactions.length > 0 &&
													selectedIds.size === transactions.length
												}
												onChange={toggleAll}
											/>
										</th>
										<th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">
											Data
										</th>
										<th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">
											Descrição
										</th>
										<th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">
											Categoria
										</th>
										<th className="text-right px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">
											Valor
										</th>
										<th className="px-4 py-3 w-20" />
									</tr>
								</thead>
								<tbody>
									{transactions.map((tx) => (
										<tr
											key={tx.id}
											className={`border-b border-zinc-800/50 hover:bg-zinc-800/30 transition-colors ${selectedIds.has(tx.id) ? "bg-indigo-900/10" : ""}`}
										>
											<td className="px-4 py-3">
												<Checkbox
													checked={selectedIds.has(tx.id)}
													onChange={() => toggleId(tx.id)}
												/>
											</td>
											<td className="px-4 py-3 text-zinc-400 whitespace-nowrap">
												{formatDate(tx.date)}
											</td>
											<td className="px-4 py-3">
												<div className="text-zinc-200 font-medium truncate max-w-xs">
													{tx.description}
												</div>
												{tx.notes && (
													<div className="text-xs text-zinc-500 truncate">
														{tx.notes}
													</div>
												)}
											</td>
											<td className="px-4 py-3">
												{tx.categoryName ? (
													<span
														className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
														style={{
															backgroundColor: `${tx.categoryColor}22`,
															color: tx.categoryColor ?? "#6366f1",
														}}
													>
														{tx.categoryName}
													</span>
												) : (
													<span className="text-zinc-600 text-xs">—</span>
												)}
											</td>
											<td
												className={`px-4 py-3 text-right font-semibold whitespace-nowrap ${tx.type === "income" ? "text-emerald-400" : "text-red-400"}`}
											>
												{tx.type === "income" ? "+" : "-"}
												{formatCurrency(tx.amount)}
											</td>
											<td className="px-4 py-3">
												<div className="flex items-center justify-end gap-1">
													<Button
														variant="ghost"
														size="icon"
														className="h-7 w-7"
														onClick={() => {
															setEditingTx(tx);
															setDialogOpen(true);
														}}
													>
														<Pencil className="h-3.5 w-3.5" />
													</Button>
													<Button
														variant="ghost"
														size="icon"
														className="h-7 w-7 text-red-400 hover:text-red-300"
														onClick={() => deleteTransaction(tx.id)}
													>
														<Trash2 className="h-3.5 w-3.5" />
													</Button>
												</div>
											</td>
										</tr>
									))}
								</tbody>
							</table>
						</div>
					)}
				</CardContent>
			</Card>

			<TransactionDialog
				open={dialogOpen}
				onClose={() => {
					setDialogOpen(false);
					setEditingTx(null);
				}}
				transaction={editingTx}
				onSaved={fetchTransactions}
				defaultMonth={month}
				defaultYear={year}
			/>

			<BulkEditDialog
				open={bulkEditOpen}
				onClose={() => setBulkEditOpen(false)}
				count={selectedIds.size}
				onSave={bulkEditSelected}
			/>
		</div>
	);
}
