"use client";

import { useState } from "react";
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
	color: string;
	type: string;
}

interface Props {
	open: boolean;
	defaultType: "expense" | "income";
	onClose: () => void;
	onCreated: (category: Category) => void;
}

const COLORS = [
	"#ef4444",
	"#f97316",
	"#f59e0b",
	"#84cc16",
	"#10b981",
	"#06b6d4",
	"#3b82f6",
	"#8b5cf6",
	"#ec4899",
	"#6b7280",
	"#14b8a6",
	"#a855f7",
	"#64748b",
	"#22c55e",
];

export function QuickAddCategoryDialog({
	open,
	defaultType,
	onClose,
	onCreated,
}: Props) {
	const [name, setName] = useState("");
	const [type, setType] = useState<string>(defaultType);
	const [color, setColor] = useState(COLORS[0]);
	const [saving, setSaving] = useState(false);

	function reset() {
		setName("");
		setType(defaultType);
		setColor(COLORS[0]);
	}

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		e.stopPropagation();
		if (!name.trim()) return;
		setSaving(true);
		const res = await fetch("/api/categories", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ name: name.trim(), color, type, icon: "circle" }),
		});
		const category = await res.json();
		setSaving(false);
		reset();
		onCreated(category);
	}

	return (
		<Dialog
			open={open}
			onOpenChange={(v) => {
				if (!v) {
					reset();
					onClose();
				}
			}}
		>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Nova categoria</DialogTitle>
				</DialogHeader>
				<form onSubmit={handleSubmit} className="space-y-4">
					<div>
						<Label>Nome</Label>
						<Input
							placeholder="Ex: Alimentação, Salário..."
							value={name}
							onChange={(e) => setName(e.target.value)}
							autoFocus
							required
							className="mt-1"
						/>
					</div>
					<div>
						<Label>Tipo</Label>
						<Select value={type} onValueChange={setType}>
							<SelectTrigger className="mt-1">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="expense">Despesa</SelectItem>
								<SelectItem value="income">Receita</SelectItem>
								<SelectItem value="both">Receita e Despesa</SelectItem>
							</SelectContent>
						</Select>
					</div>
					<div>
						<Label>Cor</Label>
						<div className="flex flex-wrap gap-2 mt-2">
							{COLORS.map((c) => (
								<button
									key={c}
									type="button"
									className={`h-7 w-7 rounded-full transition-transform hover:scale-110 ${color === c ? "ring-2 ring-white ring-offset-2 ring-offset-zinc-900 scale-110" : ""}`}
									style={{ backgroundColor: c }}
									onClick={() => setColor(c)}
								/>
							))}
						</div>
					</div>
					<DialogFooter>
						<Button
							type="button"
							variant="outline"
							onClick={() => {
								reset();
								onClose();
							}}
						>
							Cancelar
						</Button>
						<Button type="submit" disabled={saving}>
							{saving ? "Criando..." : "Criar"}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
