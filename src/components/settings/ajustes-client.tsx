"use client";

import { Bell, Mail, Save } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface Settings {
	enabled: boolean;
	emailAddress: string | null;
	daysBeforeDue: number;
	daysAfterDue: number;
	maxNotificationsPerBill: number;
	intervalDays: number;
}

const defaults: Settings = {
	enabled: true,
	emailAddress: null,
	daysBeforeDue: 3,
	daysAfterDue: 7,
	maxNotificationsPerBill: 3,
	intervalDays: 1,
};

export function AjustesClient() {
	const [settings, setSettings] = useState<Settings>(defaults);
	const [loading, setLoading] = useState(true);
	const [saving, setSaving] = useState(false);
	const [saved, setSaved] = useState(false);

	useEffect(() => {
		fetch("/api/settings")
			.then((r) => r.json())
			.then((data) => {
				setSettings({
					enabled: data.enabled ?? true,
					emailAddress: data.emailAddress ?? "",
					daysBeforeDue: data.daysBeforeDue ?? 3,
					daysAfterDue: data.daysAfterDue ?? 7,
					maxNotificationsPerBill: data.maxNotificationsPerBill ?? 3,
					intervalDays: data.intervalDays ?? 1,
				});
			})
			.finally(() => setLoading(false));
	}, []);

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		setSaving(true);
		setSaved(false);
		try {
			await fetch("/api/settings", {
				method: "PUT",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify(settings),
			});
			setSaved(true);
			setTimeout(() => setSaved(false), 3000);
		} finally {
			setSaving(false);
		}
	}

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
				<h1 className="text-2xl font-semibold text-zinc-100">Ajustes</h1>
				<p className="text-sm text-zinc-400 mt-1">
					Configure as preferências de notificação do sistema.
				</p>
			</div>

			<form onSubmit={handleSubmit} className="space-y-6">
				{/* Email Notifications */}
				<Card>
					<CardHeader>
						<CardTitle className="flex items-center gap-2 text-base">
							<Bell className="h-4 w-4 text-violet-400" />
							Notificações por e-mail
						</CardTitle>
					</CardHeader>
					<CardContent className="space-y-5">
						{/* Toggle */}
						<label className="flex items-center justify-between gap-4 cursor-pointer">
							<div>
								<p className="text-sm font-medium text-zinc-100">
									Ativar notificações
								</p>
								<p className="text-xs text-zinc-400">
									Receba e-mails sobre contas próximas do vencimento ou vencidas.
								</p>
							</div>
							<button
								type="button"
								role="switch"
								aria-checked={settings.enabled}
								onClick={() =>
									setSettings((s) => ({ ...s, enabled: !s.enabled }))
								}
								className={`relative inline-flex h-6 w-11 shrink-0 rounded-full border-2 border-transparent transition-colors focus:outline-none focus:ring-2 focus:ring-violet-500 focus:ring-offset-2 focus:ring-offset-zinc-900 ${
									settings.enabled ? "bg-violet-600" : "bg-zinc-700"
								}`}
							>
								<span
									className={`pointer-events-none block h-5 w-5 rounded-full bg-white shadow-lg ring-0 transition-transform ${
										settings.enabled ? "translate-x-5" : "translate-x-0"
									}`}
								/>
							</button>
						</label>

						<div className={settings.enabled ? "" : "opacity-40 pointer-events-none"}>
							{/* Custom email */}
							<div className="space-y-1.5">
								<Label htmlFor="emailAddress" className="flex items-center gap-1.5">
									<Mail className="h-3.5 w-3.5 text-zinc-400" />
									E-mail para notificações
								</Label>
								<Input
									id="emailAddress"
									type="email"
									placeholder="Deixe em branco para usar o e-mail da conta"
									value={settings.emailAddress ?? ""}
									onChange={(e) =>
										setSettings((s) => ({
											...s,
											emailAddress: e.target.value || null,
										}))
									}
								/>
							</div>

							<div className="grid grid-cols-2 gap-4 mt-4">
								{/* Days before */}
								<div className="space-y-1.5">
									<Label htmlFor="daysBeforeDue">Dias antes do vencimento</Label>
									<Input
										id="daysBeforeDue"
										type="number"
										min={0}
										max={30}
										value={settings.daysBeforeDue}
										onChange={(e) =>
											setSettings((s) => ({
												...s,
												daysBeforeDue: Math.max(0, Number(e.target.value)),
											}))
										}
									/>
									<p className="text-xs text-zinc-400">
										Início dos avisos antes do vencimento
									</p>
								</div>

								{/* Days after */}
								<div className="space-y-1.5">
									<Label htmlFor="daysAfterDue">Dias após o vencimento</Label>
									<Input
										id="daysAfterDue"
										type="number"
										min={0}
										max={30}
										value={settings.daysAfterDue}
										onChange={(e) =>
											setSettings((s) => ({
												...s,
												daysAfterDue: Math.max(0, Number(e.target.value)),
											}))
										}
									/>
									<p className="text-xs text-zinc-400">
										Continua avisando por quantos dias após
									</p>
								</div>

								{/* Max notifications */}
								<div className="space-y-1.5">
									<Label htmlFor="maxNotifications">Máximo de notificações</Label>
									<Input
										id="maxNotifications"
										type="number"
										min={1}
										max={10}
										value={settings.maxNotificationsPerBill}
										onChange={(e) =>
											setSettings((s) => ({
												...s,
												maxNotificationsPerBill: Math.max(
													1,
													Number(e.target.value),
												),
											}))
										}
									/>
									<p className="text-xs text-zinc-400">
										Por conta por mês
									</p>
								</div>

								{/* Interval */}
								<div className="space-y-1.5">
									<Label htmlFor="intervalDays">Intervalo entre envios</Label>
									<Input
										id="intervalDays"
										type="number"
										min={1}
										max={7}
										value={settings.intervalDays}
										onChange={(e) =>
											setSettings((s) => ({
												...s,
												intervalDays: Math.max(1, Number(e.target.value)),
											}))
										}
									/>
									<p className="text-xs text-zinc-400">Em dias</p>
								</div>
							</div>
						</div>
					</CardContent>
				</Card>

				<div className="flex items-center gap-3">
					<Button type="submit" disabled={saving} className="flex items-center gap-2">
						<Save className="h-4 w-4" />
						{saving ? "Salvando..." : "Salvar ajustes"}
					</Button>
					{saved && (
						<span className="text-sm text-emerald-400 animate-in fade-in">
							Salvo com sucesso!
						</span>
					)}
				</div>
			</form>
		</div>
	);
}
