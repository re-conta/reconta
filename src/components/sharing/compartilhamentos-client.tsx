"use client";

import { Plus, Share2, Trash2, Users } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
	Dialog,
	DialogContent,
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

interface SharedItem {
	id: number;
	targetId: string;
	targetName: string;
	targetEmail: string;
	scope: "all" | "yearly" | "monthly";
	scopeMonth: number | null;
	scopeYear: number | null;
	createdAt: string;
}

interface ReceivedItem {
	id: number;
	ownerId: string;
	ownerName: string;
	ownerEmail: string;
	scope: "all" | "yearly" | "monthly";
	scopeMonth: number | null;
	scopeYear: number | null;
	createdAt: string;
}

const MONTH_NAMES = [
	"Janeiro",
	"Fevereiro",
	"Março",
	"Abril",
	"Maio",
	"Junho",
	"Julho",
	"Agosto",
	"Setembro",
	"Outubro",
	"Novembro",
	"Dezembro",
];

function scopeLabel(item: {
	scope: "all" | "yearly" | "monthly";
	scopeMonth: number | null;
	scopeYear: number | null;
}) {
	if (item.scope === "all") return "Todos os dados";
	if (item.scope === "yearly") return `Ano ${item.scopeYear}`;
	return `${MONTH_NAMES[(item.scopeMonth ?? 1) - 1]} ${item.scopeYear}`;
}

export function CompartilhamentosClient() {
	const [shared, setShared] = useState<SharedItem[]>([]);
	const [received, setReceived] = useState<ReceivedItem[]>([]);
	const [loading, setLoading] = useState(true);
	const [dialogOpen, setDialogOpen] = useState(false);

	// Form state
	const [email, setEmail] = useState("");
	const [scope, setScope] = useState<"all" | "yearly" | "monthly">("all");
	const [scopeMonth, setScopeMonth] = useState<string>(
		String(new Date().getMonth() + 1),
	);
	const [scopeYear, setScopeYear] = useState<string>(
		String(new Date().getFullYear()),
	);
	const [saving, setSaving] = useState(false);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		Promise.all([
			fetch("/api/sharing").then((r) => r.json()),
			fetch("/api/sharing/received").then((r) => r.json()),
		])
			.then(([sharedData, receivedData]) => {
				setShared(sharedData);
				setReceived(receivedData);
			})
			.finally(() => setLoading(false));
	}, []);

	async function handleCreate(e: React.FormEvent) {
		e.preventDefault();
		setSaving(true);
		setError(null);

		try {
			const res = await fetch("/api/sharing", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					email: email.trim(),
					scope,
					scopeMonth: scope === "monthly" ? Number(scopeMonth) : undefined,
					scopeYear: scope !== "all" ? Number(scopeYear) : undefined,
				}),
			});

			const data = await res.json();
			if (!res.ok) {
				setError(data.error ?? "Erro ao compartilhar");
				return;
			}

			setShared((prev) => [
				...prev,
				{
					...data,
					targetId: data.targetId ?? "",
				},
			]);
			setDialogOpen(false);
			setEmail("");
			setScope("all");
		} finally {
			setSaving(false);
		}
	}

	async function handleRevoke(id: number) {
		await fetch(`/api/sharing/${id}`, { method: "DELETE" });
		setShared((prev) => prev.filter((s) => s.id !== id));
	}

	const currentYear = new Date().getFullYear();
	const years = Array.from({ length: 5 }, (_, i) => currentYear - 2 + i);

	if (loading) {
		return (
			<div className="flex items-center justify-center h-48 text-zinc-400">
				Carregando...
			</div>
		);
	}

	return (
		<div className="max-w-2xl space-y-6">
			<div>
				<h1 className="text-2xl font-semibold text-zinc-100">
					Compartilhamentos
				</h1>
				<p className="text-sm text-zinc-400 mt-1">
					Compartilhe seus dados financeiros com outros usuários ou visualize
					dados compartilhados com você.
				</p>
			</div>

			{/* My shares */}
			<Card>
				<CardHeader className="flex flex-row items-center justify-between pb-3">
					<CardTitle className="flex items-center gap-2 text-base">
						<Share2 className="h-4 w-4 text-violet-400" />
						Dados que estou compartilhando
					</CardTitle>
					<Button
						size="sm"
						onClick={() => {
							setError(null);
							setDialogOpen(true);
						}}
					>
						<Plus className="h-4 w-4 mr-1" />
						Compartilhar
					</Button>
				</CardHeader>
				<CardContent>
					{shared.length === 0 ? (
						<p className="text-sm text-zinc-400 text-center py-4">
							Você ainda não compartilhou dados com ninguém.
						</p>
					) : (
						<ul className="space-y-2">
							{shared.map((item) => (
								<li
									key={item.id}
									className="flex items-center justify-between gap-3 rounded-lg border border-zinc-800 px-3 py-2"
								>
									<div>
										<p className="text-sm font-medium text-zinc-100">
											{item.targetName}
										</p>
										<p className="text-xs text-zinc-400">
											{item.targetEmail} · {scopeLabel(item)}
										</p>
									</div>
									<button
										type="button"
										onClick={() => handleRevoke(item.id)}
										className="rounded p-1.5 text-zinc-500 hover:text-red-400 hover:bg-zinc-800 transition-colors"
										aria-label="Revogar acesso"
									>
										<Trash2 className="h-4 w-4" />
									</button>
								</li>
							))}
						</ul>
					)}
				</CardContent>
			</Card>

			{/* Received shares */}
			<Card>
				<CardHeader className="pb-3">
					<CardTitle className="flex items-center gap-2 text-base">
						<Users className="h-4 w-4 text-violet-400" />
						Dados compartilhados comigo
					</CardTitle>
				</CardHeader>
				<CardContent>
					{received.length === 0 ? (
						<p className="text-sm text-zinc-400 text-center py-4">
							Nenhum usuário compartilhou dados com você ainda.
						</p>
					) : (
						<ul className="space-y-2">
							{received.map((item) => (
								<li
									key={item.id}
									className="flex items-center justify-between gap-3 rounded-lg border border-zinc-800 px-3 py-2"
								>
									<div>
										<p className="text-sm font-medium text-zinc-100">
											{item.ownerName}
										</p>
										<p className="text-xs text-zinc-400">
											{item.ownerEmail} · {scopeLabel(item)}
										</p>
									</div>
									<a
										href={`/compartilhado/${item.ownerId}`}
										className="text-xs text-violet-400 hover:text-violet-300 transition-colors font-medium"
									>
										Visualizar
									</a>
								</li>
							))}
						</ul>
					)}
				</CardContent>
			</Card>

			{/* New share dialog */}
			<Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Compartilhar dados</DialogTitle>
					</DialogHeader>
					<form onSubmit={handleCreate} className="space-y-4 mt-2">
						<div className="space-y-1.5">
							<Label htmlFor="share-email">E-mail do usuário</Label>
							<Input
								id="share-email"
								type="email"
								placeholder="usuario@exemplo.com"
								value={email}
								onChange={(e) => setEmail(e.target.value)}
								required
							/>
						</div>

						<div className="space-y-1.5">
							<Label htmlFor="share-scope">Período</Label>
							<Select
								value={scope}
								onValueChange={(v) =>
									setScope(v as "all" | "yearly" | "monthly")
								}
							>
								<SelectTrigger id="share-scope">
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="all">Todos os dados</SelectItem>
									<SelectItem value="yearly">Ano específico</SelectItem>
									<SelectItem value="monthly">Mês específico</SelectItem>
								</SelectContent>
							</Select>
						</div>

						{(scope === "yearly" || scope === "monthly") && (
							<div className="space-y-1.5">
								<Label htmlFor="share-year">Ano</Label>
								<Select value={scopeYear} onValueChange={setScopeYear}>
									<SelectTrigger id="share-year">
										<SelectValue />
									</SelectTrigger>
									<SelectContent>
										{years.map((y) => (
											<SelectItem key={y} value={String(y)}>
												{y}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</div>
						)}

						{scope === "monthly" && (
							<div className="space-y-1.5">
								<Label htmlFor="share-month">Mês</Label>
								<Select value={scopeMonth} onValueChange={setScopeMonth}>
									<SelectTrigger id="share-month">
										<SelectValue />
									</SelectTrigger>
									<SelectContent>
										{MONTH_NAMES.map((name, i) => (
											<SelectItem key={name} value={String(i + 1)}>
												{name}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</div>
						)}

						{error && (
							<p className="text-sm text-red-400 bg-red-950/40 border border-red-800 rounded px-3 py-2">
								{error}
							</p>
						)}

						<div className="flex gap-2 justify-end pt-2">
							<Button
								type="button"
								variant="ghost"
								onClick={() => setDialogOpen(false)}
							>
								Cancelar
							</Button>
							<Button type="submit" disabled={saving}>
								{saving ? "Compartilhando..." : "Compartilhar"}
							</Button>
						</div>
					</form>
				</DialogContent>
			</Dialog>
		</div>
	);
}
