"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";

interface Category {
	id: number;
	name: string;
	type: string;
}

interface Account {
	id: number;
	name: string;
}

export interface BulkEditFields {
	type?: "income" | "expense";
	categoryId?: string;
	accountId?: string;
	date?: string;
}

interface Props {
	open: boolean;
	onClose: () => void;
	count: number;
	onSave: (fields: BulkEditFields) => Promise<void>;
}

export function BulkEditDialog({ open, onClose, count, onSave }: Props) {
	const [categories, setCategories] = useState<Category[]>([]);
	const [accounts, setAccounts] = useState<Account[]>([]);
	const [saving, setSaving] = useState(false);

	const [enableType, setEnableType] = useState(false);
	const [enableCategory, setEnableCategory] = useState(false);
	const [enableAccount, setEnableAccount] = useState(false);
	const [enableDate, setEnableDate] = useState(false);

	const [type, setType] = useState<"income" | "expense">("expense");
	const [categoryId, setCategoryId] = useState("");
	const [accountId, setAccountId] = useState("");
	const [date, setDate] = useState(new Date().toISOString().split("T")[0]);

	useEffect(() => {
		if (!open) return;
		Promise.all([
			fetch("/api/categories").then((r) => r.json()),
			fetch("/api/accounts").then((r) => r.json()),
		])
			.then(([cats, accs]) => {
				setCategories(cats);
				setAccounts(accs);
			})
			.catch(() => {});
		setEnableType(false);
		setEnableCategory(false);
		setEnableAccount(false);
		setEnableDate(false);
		setCategoryId("");
		setAccountId("");
	}, [open]);

	const filteredCategories = categories.filter(
		(c) => !enableType || c.type === "both" || c.type === type,
	);

	async function handleSave() {
		const fields: BulkEditFields = {};
		if (enableType) fields.type = type;
		if (enableCategory) fields.categoryId = categoryId;
		if (enableAccount) fields.accountId = accountId;
		if (enableDate) fields.date = date;
		if (Object.keys(fields).length === 0) return;
		setSaving(true);
		await onSave(fields);
		setSaving(false);
		onClose();
	}

	const hasChanges =
		enableType || enableCategory || enableAccount || enableDate;

	return (
		<Dialog open={open} onOpenChange={(v) => !v && onClose()}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>
						Editar {count} lançamento{count !== 1 ? "s" : ""}
					</DialogTitle>
				</DialogHeader>
				<p className="text-sm text-zinc-400">
					Marque os campos que deseja alterar. Os demais permanecerão
					inalterados.
				</p>
				<div className="space-y-4">
					{/* Type */}
					<div className="flex items-start gap-3">
						<Checkbox
							id="en-type"
							checked={enableType}
							onChange={(e) => setEnableType(e.target.checked)}
							className="mt-1"
						/>
						<div className="flex-1">
							<Label htmlFor="en-type" className="mb-2 block">
								Tipo
							</Label>
							<div
								className={`flex rounded-lg overflow-hidden border border-zinc-700 transition-opacity ${!enableType ? "opacity-40 pointer-events-none" : ""}`}
							>
								<button
									type="button"
									className={`flex-1 py-2 text-sm font-medium transition-colors ${type === "expense" ? "bg-red-600 text-white" : "bg-transparent text-zinc-400 hover:text-zinc-200"}`}
									onClick={() => {
										setType("expense");
										setCategoryId("");
									}}
								>
									Despesa
								</button>
								<button
									type="button"
									className={`flex-1 py-2 text-sm font-medium transition-colors ${type === "income" ? "bg-emerald-600 text-white" : "bg-transparent text-zinc-400 hover:text-zinc-200"}`}
									onClick={() => {
										setType("income");
										setCategoryId("");
									}}
								>
									Receita
								</button>
							</div>
						</div>
					</div>

					{/* Category */}
					<div className="flex items-start gap-3">
						<Checkbox
							id="en-cat"
							checked={enableCategory}
							onChange={(e) => setEnableCategory(e.target.checked)}
							className="mt-1"
						/>
						<div
							className={`flex-1 transition-opacity ${!enableCategory ? "opacity-40 pointer-events-none" : ""}`}
						>
							<Label htmlFor="en-cat" className="mb-2 block">
								Categoria
							</Label>
							<Select value={categoryId} onValueChange={setCategoryId}>
								<SelectTrigger>
									<SelectValue placeholder="Nenhuma" />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="_none">Nenhuma</SelectItem>
									{filteredCategories.map((c) => (
										<SelectItem key={c.id} value={String(c.id)}>
											{c.name}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>
					</div>

					{/* Account */}
					<div className="flex items-start gap-3">
						<Checkbox
							id="en-acc"
							checked={enableAccount}
							onChange={(e) => setEnableAccount(e.target.checked)}
							className="mt-1"
						/>
						<div
							className={`flex-1 transition-opacity ${!enableAccount ? "opacity-40 pointer-events-none" : ""}`}
						>
							<Label htmlFor="en-acc" className="mb-2 block">
								Conta
							</Label>
							<Select value={accountId} onValueChange={setAccountId}>
								<SelectTrigger>
									<SelectValue placeholder="Nenhuma" />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="_none">Nenhuma</SelectItem>
									{accounts.map((a) => (
										<SelectItem key={a.id} value={String(a.id)}>
											{a.name}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>
					</div>

					{/* Date */}
					<div className="flex items-start gap-3">
						<Checkbox
							id="en-date"
							checked={enableDate}
							onChange={(e) => setEnableDate(e.target.checked)}
							className="mt-1"
						/>
						<div
							className={`flex-1 transition-opacity ${!enableDate ? "opacity-40 pointer-events-none" : ""}`}
						>
							<Label htmlFor="en-date" className="mb-2 block">
								Data
							</Label>
							<Input
								type="date"
								value={date}
								onChange={(e) => setDate(e.target.value)}
							/>
						</div>
					</div>
				</div>

				<DialogFooter>
					<Button variant="outline" onClick={onClose}>
						Cancelar
					</Button>
					<Button onClick={handleSave} disabled={saving || !hasChanges}>
						{saving ? "Salvando..." : "Salvar alterações"}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
