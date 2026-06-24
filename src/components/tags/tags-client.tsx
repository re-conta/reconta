"use client";

import { Pencil, Plus, Trash2 } from "lucide-react";
import { useEffect, useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { useSharedOwner } from "@/components/layout/shared-owner-context";
import { Card, CardContent } from "@/components/ui/card";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface Tag {
	id: number;
	name: string;
	color: string;
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

export function TagsClient() {
	const shared = useSharedOwner();
	const apiBase = shared?.apiBase ?? "/api";
	const readOnly = !!shared;
	const [tags, setTags] = useState<Tag[]>([]);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [editing, setEditing] = useState<Tag | null>(null);
	const [form, setForm] = useState({ name: "", color: COLORS[0] });
	const [saving, setSaving] = useState(false);

	const fetchTags = useCallback(() => {
		fetch(`${apiBase}/tags`)
			.then((r) => r.json())
			.then(setTags);
	}, [apiBase]);

	useEffect(() => {
		fetchTags();
	}, [fetchTags]);

	function openNew() {
		setEditing(null);
		setForm({ name: "", color: COLORS[0] });
		setDialogOpen(true);
	}

	function openEdit(tag: Tag) {
		setEditing(tag);
		setForm({ name: tag.name, color: tag.color });
		setDialogOpen(true);
	}

	async function handleDelete(id: number) {
		if (!confirm("Deseja excluir esta tag?")) return;
		await fetch(`/api/tags/${id}`, { method: "DELETE" });
		fetchTags();
	}

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		setSaving(true);
		const url = editing ? `/api/tags/${editing.id}` : "/api/tags";
		const method = editing ? "PUT" : "POST";
		await fetch(url, {
			method,
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(form),
		});
		setSaving(false);
		fetchTags();
		setDialogOpen(false);
	}

	return (
		<div className="max-w-lg space-y-6">
			{!readOnly && (
				<Button onClick={openNew}>
					<Plus className="h-4 w-4" />
					Nova tag
				</Button>
			)}

			<Card>
				<CardContent className="pt-5">
					<div className="space-y-1">
						{tags.map((tag) => (
							<div
								key={tag.id}
								className="flex items-center gap-3 rounded-lg p-3 bg-zinc-900 border border-zinc-800 hover:border-zinc-700 transition-colors"
							>
								<span
									className="h-4 w-4 rounded-full shrink-0"
									style={{ backgroundColor: tag.color }}
								/>
								<span className="flex-1 text-zinc-200 text-sm">{tag.name}</span>
								{!readOnly && (
									<div className="flex gap-1">
										<Button
											variant="ghost"
											size="icon"
											className="h-7 w-7"
											onClick={() => openEdit(tag)}
										>
											<Pencil className="h-3.5 w-3.5" />
										</Button>
										<Button
											variant="ghost"
											size="icon"
											className="h-7 w-7 text-red-400 hover:text-red-300"
											onClick={() => handleDelete(tag.id)}
										>
											<Trash2 className="h-3.5 w-3.5" />
										</Button>
									</div>
								)}
							</div>
						))}
						{tags.length === 0 && (
							<p className="text-zinc-500 text-sm text-center py-4">
								Nenhuma tag cadastrada.
							</p>
						)}
					</div>
				</CardContent>
			</Card>

			{!readOnly && (
				<Dialog
					open={dialogOpen}
					onOpenChange={(v) => !v && setDialogOpen(false)}
				>
					<DialogContent>
						<DialogHeader>
							<DialogTitle>{editing ? "Editar tag" : "Nova tag"}</DialogTitle>
						</DialogHeader>
						<form onSubmit={handleSubmit} className="space-y-4">
							<div>
								<Label>Nome</Label>
								<Input
									placeholder="Ex: Viagem, Reembolsável..."
									value={form.name}
									onChange={(e) =>
										setForm((f) => ({ ...f, name: e.target.value }))
									}
									required
									className="mt-1"
								/>
							</div>
							<div>
								<Label>Cor</Label>
								<div className="flex flex-wrap gap-2 mt-2">
									{COLORS.map((c) => (
										<button
											key={c}
											type="button"
											className={`h-7 w-7 rounded-full transition-transform hover:scale-110 ${form.color === c ? "ring-2 ring-white ring-offset-2 ring-offset-zinc-900 scale-110" : ""}`}
											style={{ backgroundColor: c }}
											onClick={() => setForm((f) => ({ ...f, color: c }))}
										/>
									))}
								</div>
							</div>
							<DialogFooter>
								<Button
									type="button"
									variant="outline"
									onClick={() => setDialogOpen(false)}
								>
									Cancelar
								</Button>
								<Button type="submit" disabled={saving}>
									{saving ? "Salvando..." : editing ? "Salvar" : "Criar"}
								</Button>
							</DialogFooter>
						</form>
					</DialogContent>
				</Dialog>
			)}
		</div>
	);
}
