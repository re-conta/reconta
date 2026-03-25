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

interface Category {
	id: number;
	name: string;
}

interface Bill {
	id: number;
	name: string;
	amount: number;
	dueDay: number;
	categoryId: number | null;
}

interface Props {
	open: boolean;
	onClose: () => void;
	bill: Bill | null;
	onSaved: () => void;
}

export function BillDialog({ open, onClose, bill, onSaved }: Props) {
	const [categories, setCategories] = useState<Category[]>([]);
	const [form, setForm] = useState({
		name: "",
		amount: "",
		dueDay: "1",
		categoryId: "",
	});
	const [saving, setSaving] = useState(false);

	useEffect(() => {
		fetch("/api/categories")
			.then((r) => {
				if (!r.ok) throw new Error(r.statusText);
				return r.json();
			})
			.then(setCategories)
			.catch(() => {});
	}, []);

	useEffect(() => {
		if (bill) {
			setForm({
				name: bill.name,
				amount: String(bill.amount),
				dueDay: String(bill.dueDay),
				categoryId: bill.categoryId ? String(bill.categoryId) : "",
			});
		} else {
			setForm({ name: "", amount: "", dueDay: "1", categoryId: "" });
		}
	}, [bill]);

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		setSaving(true);
		const body = {
			name: form.name,
			amount: Number(form.amount),
			dueDay: Number(form.dueDay),
			categoryId: form.categoryId || null,
		};
		const url = bill ? `/api/bills/${bill.id}` : "/api/bills";
		const method = bill ? "PUT" : "POST";
		await fetch(url, {
			method,
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(body),
		});
		setSaving(false);
		onSaved();
		onClose();
	}

	return (
		<Dialog open={open} onOpenChange={(v) => !v && onClose()}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>
						{bill ? "Editar conta fixa" : "Nova conta fixa"}
					</DialogTitle>
				</DialogHeader>
				<form onSubmit={handleSubmit} className="space-y-4">
					<div>
						<Label htmlFor="bill-name">Nome</Label>
						<Input
							id="bill-name"
							placeholder="Ex: Aluguel, Internet, Condomínio..."
							value={form.name}
							onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
							required
							className="mt-1"
						/>
					</div>
					<div className="grid grid-cols-2 gap-3">
						<div>
							<Label htmlFor="bill-amount">Valor (R$)</Label>
							<Input
								id="bill-amount"
								type="number"
								step="0.01"
								min="0.01"
								placeholder="0,00"
								value={form.amount}
								onChange={(e) =>
									setForm((f) => ({ ...f, amount: e.target.value }))
								}
								required
								className="mt-1"
							/>
						</div>
						<div>
							<Label htmlFor="bill-due">Dia de vencimento</Label>
							<Input
								id="bill-due"
								type="number"
								min="1"
								max="31"
								placeholder="Ex: 10"
								value={form.dueDay}
								onChange={(e) =>
									setForm((f) => ({ ...f, dueDay: e.target.value }))
								}
								required
								className="mt-1"
							/>
						</div>
					</div>
					<div>
						<Label>Categoria</Label>
						<Select
							value={form.categoryId}
							onValueChange={(v) => setForm((f) => ({ ...f, categoryId: v }))}
						>
							<SelectTrigger className="mt-1">
								<SelectValue placeholder="Selecionar..." />
							</SelectTrigger>
							<SelectContent>
								{categories.map((c) => (
									<SelectItem key={c.id} value={String(c.id)}>
										{c.name}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>
					<DialogFooter>
						<Button type="button" variant="outline" onClick={onClose}>
							Cancelar
						</Button>
						<Button type="submit" disabled={saving}>
							{saving ? "Salvando..." : bill ? "Salvar" : "Adicionar"}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
