"use client";

import { Check, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";

interface Category {
	id: number;
	name: string;
	color: string;
	type: string;
}

interface Account {
	id: number;
	name: string;
}

interface Transaction {
	id: number;
	date: string;
	description: string;
	amount: number;
	type: string;
	categoryId: number | null;
	accountId: number | null;
	notes: string | null;
}

interface Props {
	transaction?: Transaction | null;
	defaultMonth: number;
	defaultYear: number;
	onSaved: () => void;
	onCancel?: () => void;
}

function defaultDate(month: number, year: number) {
	const today = new Date();
	const day = Math.min(today.getDate(), new Date(year, month, 0).getDate());
	return new Date(year, month - 1, day).toISOString().split("T")[0];
}

const inputCls =
	"h-9 rounded-lg border border-zinc-700 bg-zinc-800 px-3 text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-indigo-500";

const selectCls =
	"h-9 rounded-lg border border-zinc-700 bg-zinc-800 pl-2.5 pr-7 text-sm text-zinc-100 appearance-none focus:outline-none focus:ring-2 focus:ring-indigo-500";

const chevronStyle = {
	colorScheme: "dark" as const,
	backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' fill='none' stroke='%2371717a' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m4 6 4 4 4-4'/%3E%3C/svg%3E")`,
	backgroundPosition: "right 0.5rem center",
	backgroundSize: "1rem",
	backgroundRepeat: "no-repeat",
};

export function InlineTransactionForm({
	transaction,
	defaultMonth,
	defaultYear,
	onSaved,
	onCancel,
}: Props) {
	const [categories, setCategories] = useState<Category[]>([]);
	const [accounts, setAccounts] = useState<Account[]>([]);
	const [saving, setSaving] = useState(false);
	const [type, setType] = useState<"expense" | "income">("expense");
	const [date, setDate] = useState(() =>
		defaultDate(defaultMonth, defaultYear),
	);
	const [description, setDescription] = useState("");
	const [amount, setAmount] = useState("");
	const [categoryId, setCategoryId] = useState("");
	const [accountId, setAccountId] = useState("");

	const descRef = useRef<HTMLInputElement>(null);
	const formRef = useRef<HTMLFormElement>(null);

	useEffect(() => {
		Promise.all([
			fetch("/api/categories").then((r) => r.json()),
			fetch("/api/accounts").then((r) => r.json()),
		]).then(([cats, accs]) => {
			setCategories(Array.isArray(cats) ? cats : []);
			setAccounts(Array.isArray(accs) ? accs : []);
		});
	}, []);

	useEffect(() => {
		if (transaction) {
			setType(transaction.type as "expense" | "income");
			setDate(transaction.date);
			setDescription(transaction.description);
			setAmount(String(transaction.amount));
			setCategoryId(
				transaction.categoryId ? String(transaction.categoryId) : "",
			);
			setAccountId(transaction.accountId ? String(transaction.accountId) : "");
			formRef.current?.scrollIntoView({ behavior: "smooth", block: "nearest" });
			setTimeout(() => descRef.current?.focus(), 100);
		}
	}, [transaction]);

	useEffect(() => {
		if (!transaction) {
			setDate(defaultDate(defaultMonth, defaultYear));
		}
	}, [transaction, defaultMonth, defaultYear]);

	function reset() {
		setType("expense");
		setDate(defaultDate(defaultMonth, defaultYear));
		setDescription("");
		setAmount("");
		setCategoryId("");
		setAccountId("");
		setTimeout(() => descRef.current?.focus(), 50);
	}

	const filteredCategories = categories.filter(
		(c) => c.type === "both" || c.type === type,
	);

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		if (!description.trim() || !amount) return;

		setSaving(true);
		const body = {
			date,
			description: description.trim(),
			amount: Number(amount),
			type,
			categoryId: categoryId || null,
			accountId: accountId || null,
			notes: transaction?.notes ?? null,
		};

		const url = transaction
			? `/api/transactions/${transaction.id}`
			: "/api/transactions";
		const method = transaction ? "PUT" : "POST";

		await fetch(url, {
			method,
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(body),
		});

		setSaving(false);
		onSaved();
		if (transaction) {
			onCancel?.();
		} else {
			reset();
		}
	}

	const isEditing = !!transaction;

	return (
		<form
			ref={formRef}
			onSubmit={handleSubmit}
			className={`rounded-lg border p-3 transition-colors ${
				isEditing
					? "border-indigo-600/50 bg-indigo-950/20"
					: "border-zinc-700/60 bg-zinc-900/60"
			}`}
		>
			<div className="flex flex-wrap gap-2 items-center">
				{/* Type toggle */}
				<button
					type="button"
					onClick={() => {
						setType((t) => (t === "expense" ? "income" : "expense"));
						setCategoryId("");
					}}
					className={`h-9 px-3 shrink-0 rounded-lg text-sm font-medium transition-colors cursor-pointer ${
						type === "expense"
							? "bg-red-500/15 text-red-400 border border-red-700/40"
							: "bg-emerald-500/15 text-emerald-400 border border-emerald-700/40"
					}`}
				>
					{type === "expense" ? "Despesa" : "Receita"}
				</button>

				{/* Date */}
				<input
					type="date"
					value={date}
					onChange={(e) => setDate(e.target.value)}
					enterKeyHint="next"
					className={`${inputCls} w-[8.5rem] shrink-0`}
					style={{ colorScheme: "dark" }}
					required
				/>

				{/* Description */}
				<input
					ref={descRef}
					type="text"
					placeholder="Descrição..."
					value={description}
					onChange={(e) => setDescription(e.target.value)}
					enterKeyHint="next"
					className={`${inputCls} min-w-[10rem] flex-1`}
					required
				/>

				{/* Amount */}
				<input
					type="number"
					step="0.01"
					min="0.01"
					placeholder="Valor"
					value={amount}
					onChange={(e) => setAmount(e.target.value)}
					enterKeyHint="next"
					inputMode="decimal"
					className={`${inputCls} w-24 shrink-0`}
					required
				/>

				{/* Category */}
				<select
					value={categoryId}
					onChange={(e) => setCategoryId(e.target.value)}
					className={`${selectCls} w-32 shrink-0`}
					style={chevronStyle}
				>
					<option value="">Categoria</option>
					{filteredCategories.map((c) => (
						<option key={c.id} value={String(c.id)}>
							{c.name}
						</option>
					))}
				</select>

				{/* Account */}
				<select
					value={accountId}
					onChange={(e) => setAccountId(e.target.value)}
					className={`${selectCls} w-32 shrink-0`}
					style={chevronStyle}
				>
					<option value="">Conta</option>
					{accounts.map((a) => (
						<option key={a.id} value={String(a.id)}>
							{a.name}
						</option>
					))}
				</select>

				{/* Actions */}
				<div className="flex gap-1 shrink-0">
					<Button
						type="submit"
						size="icon"
						className="h-9 w-9"
						disabled={saving}
					>
						<Check className="h-4 w-4" />
					</Button>
					{isEditing && (
						<Button
							type="button"
							size="icon"
							variant="ghost"
							className="h-9 w-9"
							onClick={onCancel}
						>
							<X className="h-4 w-4" />
						</Button>
					)}
				</div>
			</div>

			{isEditing && (
				<p className="text-xs text-indigo-400/70 mt-2">
					Editando lançamento —{" "}
					<button
						type="button"
						className="underline cursor-pointer hover:text-indigo-300"
						onClick={onCancel}
					>
						cancelar
					</button>
				</p>
			)}
		</form>
	);
}
