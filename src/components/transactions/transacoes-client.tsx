"use client";

import {
	ChevronDown,
	ChevronLeft,
	ChevronRight,
	Pencil,
	Plus,
	Scale,
	Search,
	Sparkles,
	Trash2,
	TrendingDown,
	TrendingUp,
	Wallet,
} from "lucide-react";
import { useEffect, useRef, useState, useCallback, useId } from "react";
import { Button } from "@/components/ui/button";
import { useMonthContext } from "@/components/layout/month-context";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { formatCurrency, formatDate, formatMonth } from "@/lib/utils";
import { InlineTransactionForm } from "./inline-transaction-form";
import { TransactionDialog } from "./transaction-dialog";
import { BulkEditDialog, type BulkEditFields } from "./bulk-edit-dialog";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useSharedOwner } from "@/components/layout/shared-owner-context";

interface Tag {
	id: number;
	name: string;
	color: string;
}

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
	bank: string | null;
	pixBeneficiary: string | null;
	tags: Tag[];
}

interface Totals {
	income: number;
	expense: number;
	balance: number;
	count: number;
}

export function TransacoesClient() {
	const { month, year, setPeriod } = useMonthContext();
	const shared = useSharedOwner();
	const apiBase = shared ? shared.apiBase : "/api";
	const readOnly = !!shared;
	const [search, setSearch] = useState("");
	const [typeFilter, setTypeFilter] = useState<"all" | "income" | "expense">(
		"all",
	);
	const [tagFilter, setTagFilter] = useState<number | null>(null);
	const [allTags, setAllTags] = useState<Tag[]>([]);
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
	const [inlineEditingTx, setInlineEditingTx] = useState<Transaction | null>(
		null,
	);
	const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
	const [bulkEditOpen, setBulkEditOpen] = useState(false);
	const [autoCategorizing, setAutoCategorizing] = useState(false);

	const [openingBalance, setOpeningBalance] = useState(0);
	const [openingBalanceInput, setOpeningBalanceInput] = useState("");
	const [editingOpeningBalance, setEditingOpeningBalance] = useState(false);
	const openingBalanceInputId = useId();
	const selectAllMobileId = useId();

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
		if (tagFilter) params.set("tagId", String(tagFilter));
		if (search) params.set("search", search);

		fetch(`${apiBase}/transactions?${params}`)
			.then((r) => r.json())
			.then((d) => {
				setTransactions(d.data ?? []);
				setTotals(d.totals ?? { income: 0, expense: 0, balance: 0, count: 0 });
				setLoading(false);
			});
	}, [month, year, typeFilter, tagFilter, search, apiBase]);

	const fetchOpeningBalance = useCallback(() => {
		fetch(`/api/transactions/opening-balance?month=${month}&year=${year}`)
			.then((r) => r.json())
			.then((d) => setOpeningBalance(d.amount ?? 0));
	}, [month, year]);

	useEffect(() => {
		fetchTransactions();
	}, [fetchTransactions]);

	useEffect(() => {
		fetch(`${apiBase}/tags`)
			.then((r) => r.json())
			.then((d) => setAllTags(Array.isArray(d) ? d : []))
			.catch(() => {});
	}, [apiBase]);

	useEffect(() => {
		fetchOpeningBalance();
	}, [fetchOpeningBalance]);

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

	async function saveOpeningBalance() {
		const value = Number.parseFloat(openingBalanceInput.replace(",", "."));
		if (Number.isNaN(value)) {
			setEditingOpeningBalance(false);
			return;
		}
		await fetch("/api/transactions/opening-balance", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ month, year, amount: value }),
		});
		setOpeningBalance(value);
		setEditingOpeningBalance(false);
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

	async function autoCategorize() {
		setAutoCategorizing(true);
		try {
			const res = await fetch("/api/transactions/auto-categorize", {
				method: "POST",
			});
			const data = await res.json();
			fetchTransactions();
			alert(
				data.updated > 0
					? `${data.updated} lançamento${data.updated !== 1 ? "s" : ""} categorizado${data.updated !== 1 ? "s" : ""} automaticamente.`
					: "Nenhum lançamento correspondeu aos padrões cadastrados.",
			);
		} finally {
			setAutoCategorizing(false);
		}
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

	const finalBalance = openingBalance + totals.balance;

	return (
		<div className="space-y-5">
			{/* Controls */}
			<div className="flex flex-wrap items-center justify-between gap-3">
				<div className="flex items-center gap-1 rounded-full border border-zinc-800 bg-zinc-900/70 p-1 shadow-sm">
					<Button
						variant="ghost"
						size="icon"
						className="h-8 w-8 rounded-full"
						onClick={prevMonth}
					>
						<ChevronLeft className="h-4 w-4" />
					</Button>
					<span className="font-nunito font-semibold capitalize min-w-32 text-center text-sm text-zinc-100">
						{formatMonth(month, year)}
					</span>
					<Button
						variant="ghost"
						size="icon"
						className="h-8 w-8 rounded-full"
						onClick={nextMonth}
						disabled={isCurrentMonth}
					>
						<ChevronRight className="h-4 w-4" />
					</Button>
				</div>
				{!readOnly && (
					<div className="flex items-center gap-2">
						<Button
							size="sm"
							variant="outline"
							onClick={autoCategorize}
							disabled={autoCategorizing}
						>
							<Sparkles className="h-4 w-4" />
							<span className="hidden sm:inline">
								{autoCategorizing ? "Categorizando..." : "Auto-categorizar"}
							</span>
						</Button>
						<Button
							size="sm"
							onClick={() => {
								setEditingTx(null);
								setDialogOpen(true);
							}}
							className="shadow-md shadow-indigo-950/40"
						>
							<Plus className="h-4 w-4" />
							<span className="hidden sm:inline">Novo lançamento</span>
						</Button>
					</div>
				)}
			</div>

			{/* Totals */}
			<div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
				{!readOnly && (
					<div className="relative overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-900/60 p-4 transition-colors hover:border-zinc-700">
						<div className="flex items-center justify-between mb-2.5">
							<p className="text-xs font-medium text-zinc-400 uppercase tracking-wide">
								Saldo inicial
							</p>
							<span className="flex h-7 w-7 items-center justify-center rounded-full bg-zinc-800 text-zinc-400">
								<Wallet className="h-3.5 w-3.5" />
							</span>
						</div>
						{editingOpeningBalance ? (
							<Input
								id={openingBalanceInputId}
								autoFocus
								value={openingBalanceInput}
								onChange={(e) => setOpeningBalanceInput(e.target.value)}
								onBlur={saveOpeningBalance}
								onKeyDown={(e) => {
									if (e.key === "Enter") saveOpeningBalance();
									if (e.key === "Escape") setEditingOpeningBalance(false);
								}}
								className="h-7 text-sm w-full"
								placeholder="0,00"
							/>
						) : (
							<button
								type="button"
								className="group flex items-center gap-1.5 text-lg sm:text-xl font-bold font-nunito text-zinc-200 hover:text-white cursor-pointer text-left w-full"
								onClick={() => {
									setOpeningBalanceInput(String(openingBalance));
									setEditingOpeningBalance(true);
								}}
								title="Clique para editar o saldo inicial"
							>
								{formatCurrency(openingBalance)}
								<Pencil className="h-3 w-3 text-zinc-600 group-hover:text-zinc-400 transition-colors" />
							</button>
						)}
					</div>
				)}
				<div className="relative overflow-hidden rounded-2xl border border-emerald-800/30 bg-gradient-to-br from-emerald-900/25 to-emerald-950/10 p-4 transition-colors hover:border-emerald-700/40">
					<div className="flex items-center justify-between mb-2.5">
						<p className="text-xs font-medium text-emerald-300/70 uppercase tracking-wide">
							Receitas
						</p>
						<span className="flex h-7 w-7 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-400">
							<TrendingUp className="h-3.5 w-3.5" />
						</span>
					</div>
					<p className="text-lg sm:text-xl font-bold font-nunito text-emerald-400">
						{formatCurrency(totals.income)}
					</p>
				</div>
				<div className="relative overflow-hidden rounded-2xl border border-red-800/30 bg-gradient-to-br from-red-900/25 to-red-950/10 p-4 transition-colors hover:border-red-700/40">
					<div className="flex items-center justify-between mb-2.5">
						<p className="text-xs font-medium text-red-300/70 uppercase tracking-wide">
							Despesas
						</p>
						<span className="flex h-7 w-7 items-center justify-center rounded-full bg-red-500/15 text-red-400">
							<TrendingDown className="h-3.5 w-3.5" />
						</span>
					</div>
					<p className="text-lg sm:text-xl font-bold font-nunito text-red-400">
						{formatCurrency(totals.expense)}
					</p>
				</div>
				<div
					className={`relative overflow-hidden rounded-2xl border p-4 transition-colors ${
						finalBalance >= 0
							? "border-indigo-800/30 bg-gradient-to-br from-indigo-900/25 to-indigo-950/10 hover:border-indigo-700/40"
							: "border-red-800/30 bg-gradient-to-br from-red-900/25 to-red-950/10 hover:border-red-700/40"
					}`}
				>
					<div className="flex items-center justify-between mb-2.5">
						<p
							className={`text-xs font-medium uppercase tracking-wide ${finalBalance >= 0 ? "text-indigo-300/70" : "text-red-300/70"}`}
						>
							Saldo final
						</p>
						<span
							className={`flex h-7 w-7 items-center justify-center rounded-full ${finalBalance >= 0 ? "bg-indigo-500/15 text-indigo-400" : "bg-red-500/15 text-red-400"}`}
						>
							<Scale className="h-3.5 w-3.5" />
						</span>
					</div>
					<p
						className={`text-lg sm:text-xl font-bold font-nunito ${finalBalance >= 0 ? "text-indigo-400" : "text-red-400"}`}
					>
						{formatCurrency(finalBalance)}
					</p>
				</div>
			</div>

			{/* Filter bar */}
			<div className="flex flex-wrap items-center gap-2 rounded-2xl border border-zinc-800 bg-zinc-900/50 p-2">
				<div className="relative flex-1 min-w-40">
					<Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
					<Input
						placeholder="Buscar lançamentos..."
						value={search}
						onChange={(e) => setSearch(e.target.value)}
						className="pl-9 bg-zinc-900/80 border-zinc-700/60"
					/>
				</div>
				<div className="flex gap-1 rounded-lg bg-zinc-800/60 p-1">
					{(["all", "income", "expense"] as const).map((t) => (
						<button
							key={t}
							type="button"
							onClick={() => setTypeFilter(t)}
							className={`rounded-md px-3 h-7 text-xs font-medium transition-colors cursor-pointer ${
								typeFilter === t
									? "bg-indigo-600 text-white shadow-sm"
									: "text-zinc-400 hover:text-zinc-200"
							}`}
						>
							{t === "all" ? "Todos" : t === "income" ? "Receitas" : "Despesas"}
						</button>
					))}
				</div>
				{allTags.length > 0 && (
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button variant="outline" size="sm">
								{tagFilter ? (
									<span
										className="h-2.5 w-2.5 rounded-full"
										style={{
											backgroundColor: allTags.find((t) => t.id === tagFilter)
												?.color,
										}}
									/>
								) : null}
								<span>
									{tagFilter
										? allTags.find((t) => t.id === tagFilter)?.name
										: "Tag"}
								</span>
								<ChevronDown className="h-3.5 w-3.5" />
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="start">
							<DropdownMenuItem onSelect={() => setTagFilter(null)}>
								Todas
							</DropdownMenuItem>
							<DropdownMenuSeparator />
							{allTags.map((t) => (
								<DropdownMenuItem
									key={t.id}
									onSelect={() => setTagFilter(t.id)}
								>
									<span
										className="h-2.5 w-2.5 rounded-full mr-2"
										style={{ backgroundColor: t.color }}
									/>
									{t.name}
								</DropdownMenuItem>
							))}
						</DropdownMenuContent>
					</DropdownMenu>
				)}
				{!readOnly && (
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button variant="outline" size="sm">
								<Trash2 className="h-4 w-4 text-red-400" />
								<span className="hidden sm:inline">Apagar em massa</span>
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
				)}
			</div>

			{/* Bulk action bar */}
			{!readOnly && selectedIds.size > 0 && (
				<div className="flex flex-wrap items-center gap-3 px-4 py-2.5 rounded-xl bg-indigo-900/30 border border-indigo-700/40 text-sm">
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

			{/* Inline form */}
			{!readOnly && (
				<InlineTransactionForm
					transaction={inlineEditingTx}
					defaultMonth={month}
					defaultYear={year}
					onSaved={() => {
						setInlineEditingTx(null);
						fetchTransactions();
					}}
					onCancel={() => setInlineEditingTx(null)}
				/>
			)}

			{/* Loading / empty states */}
			{loading ? (
				<Card>
					<CardContent className="p-10 text-center text-zinc-500 text-sm">
						Carregando...
					</CardContent>
				</Card>
			) : transactions.length === 0 ? (
				<Card>
					<CardContent className="p-10 text-center text-zinc-500 text-sm">
						Nenhum lançamento encontrado.
					</CardContent>
				</Card>
			) : (
				<>
					{/* Mobile: card list */}
					<div className="space-y-2 md:hidden">
						{!readOnly && (
							<label
								htmlFor={selectAllMobileId}
								className="flex items-center gap-2 px-1 text-xs text-zinc-400"
							>
								<Checkbox
									id={selectAllMobileId}
									ref={selectAllRef}
									checked={
										transactions.length > 0 &&
										selectedIds.size === transactions.length
									}
									onChange={toggleAll}
								/>
								Selecionar todos
							</label>
						)}
						{transactions.map((tx) => (
							<Card
								key={tx.id}
								className={`transition-colors ${selectedIds.has(tx.id) ? "border-indigo-700/50 bg-indigo-900/10" : ""}`}
							>
								<CardContent className="p-3.5">
									<div className="flex items-start gap-3">
										{!readOnly && (
											<Checkbox
												checked={selectedIds.has(tx.id)}
												onChange={() => toggleId(tx.id)}
												className="mt-1"
											/>
										)}
										<div className="flex-1 min-w-0">
											<button
												type="button"
												className="block w-full text-left"
												onClick={
													readOnly ? undefined : () => setInlineEditingTx(tx)
												}
												disabled={readOnly}
											>
												<div className="flex items-center justify-between gap-2">
													<p className="font-medium text-zinc-200 truncate">
														{tx.description}
													</p>
													<p
														className={`font-semibold whitespace-nowrap font-nunito ${tx.type === "income" ? "text-emerald-400" : "text-red-400"}`}
													>
														{tx.type === "income" ? "+" : "-"}
														{formatCurrency(tx.amount)}
													</p>
												</div>
											</button>
											<div className="flex items-center justify-between gap-2 mt-1.5">
												<div className="flex items-center gap-2 min-w-0">
													<span className="text-xs text-zinc-500 whitespace-nowrap">
														{formatDate(tx.date)}
													</span>
													{tx.categoryName ? (
														<span
															className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium truncate"
															style={{
																backgroundColor: `${tx.categoryColor}22`,
																color: tx.categoryColor ?? "#6366f1",
															}}
														>
															{tx.categoryName}
														</span>
													) : null}
													{tx.tags.map((tag) => (
														<span
															key={tag.id}
															className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium truncate"
															style={{
																backgroundColor: `${tag.color}22`,
																color: tag.color,
															}}
														>
															{tag.name}
														</span>
													))}
												</div>
												{!readOnly && (
													<Button
														variant="ghost"
														size="icon"
														className="h-7 w-7 text-red-400 hover:text-red-300 shrink-0"
														onClick={() => deleteTransaction(tx.id)}
													>
														<Trash2 className="h-3.5 w-3.5" />
													</Button>
												)}
											</div>
											{(tx.pixBeneficiary || tx.bank || tx.notes) && (
												<div className="mt-1 text-xs text-zinc-500 truncate">
													{tx.pixBeneficiary && (
														<span className="text-indigo-400/70">
															Beneficiário: {tx.pixBeneficiary}
														</span>
													)}
													{tx.bank && <span> · {tx.bank}</span>}
													{tx.notes && <span> · {tx.notes}</span>}
												</div>
											)}
										</div>
									</div>
								</CardContent>
							</Card>
						))}
					</div>

					{/* Desktop: table */}
					<Card className="hidden md:block">
						<CardContent className="p-0">
							<div className="overflow-x-auto">
								<table className="w-full text-sm">
									<thead>
										<tr className="border-b border-zinc-800">
											{!readOnly && (
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
											)}
											<th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">
												Data
											</th>
											<th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">
												Descrição
											</th>
											<th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">
												Categoria
											</th>
											<th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">
												Tags
											</th>
											<th className="text-right px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">
												Valor
											</th>
											{!readOnly && <th className="px-4 py-3 w-20" />}
										</tr>
									</thead>
									<tbody>
										{transactions.map((tx) => (
											<tr
												key={tx.id}
												className={`border-b border-zinc-800/50 hover:bg-zinc-800/30 transition-colors ${selectedIds.has(tx.id) ? "bg-indigo-900/10" : ""}`}
											>
												{!readOnly && (
													<td className="px-4 py-3">
														<Checkbox
															checked={selectedIds.has(tx.id)}
															onChange={() => toggleId(tx.id)}
														/>
													</td>
												)}
												<td className="px-4 py-3 text-zinc-400 whitespace-nowrap">
													{formatDate(tx.date)}
												</td>
												<td
													className={`group px-4 py-3 ${!readOnly ? "cursor-pointer" : ""}`}
													onClick={
														readOnly ? undefined : () => setInlineEditingTx(tx)
													}
													onKeyDown={
														readOnly
															? undefined
															: (e) => {
																	if (e.key === "Enter" || e.key === " ") {
																		e.preventDefault();
																		setInlineEditingTx(tx);
																	}
																}
													}
													role={readOnly ? undefined : "button"}
													tabIndex={readOnly ? undefined : 0}
													title={readOnly ? undefined : "Clique para editar"}
												>
													<div className="flex items-center gap-1.5 text-zinc-200 font-medium truncate">
														{tx.description}
														{!readOnly && (
															<Pencil className="h-3 w-3 text-zinc-600 group-hover:text-zinc-400 transition-colors shrink-0" />
														)}
													</div>
													{tx.pixBeneficiary && (
														<div className="text-xs text-indigo-400/70 truncate">
															Beneficiário: {tx.pixBeneficiary}
														</div>
													)}
													{tx.bank && (
														<div className="text-xs text-zinc-500 truncate">
															{tx.bank}
														</div>
													)}
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
												<td className="px-4 py-3">
													<div className="flex flex-wrap gap-1 max-w-44">
														{tx.tags.map((tag) => (
															<span
																key={tag.id}
																className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium truncate"
																style={{
																	backgroundColor: `${tag.color}22`,
																	color: tag.color,
																}}
															>
																{tag.name}
															</span>
														))}
													</div>
												</td>
												<td
													className={`px-4 py-3 text-right font-semibold whitespace-nowrap ${tx.type === "income" ? "text-emerald-400" : "text-red-400"}`}
												>
													{tx.type === "income" ? "+" : "-"}
													{formatCurrency(tx.amount)}
												</td>
												{!readOnly && (
													<td className="px-4 py-3">
														<div className="flex items-center justify-end gap-1">
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
												)}
											</tr>
										))}
									</tbody>
								</table>
							</div>
						</CardContent>
					</Card>
				</>
			)}

			{!readOnly && (
				<>
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
				</>
			)}
		</div>
	);
}
