"use client";

import { AlertTriangle, Clock } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { useSession } from "@/lib/auth-client";
import { formatCurrency } from "@/lib/utils";

interface OverdueBill {
	id: number;
	name: string;
	amount: number;
	dueDay: number;
	daysUntil: number;
	isOverdue: boolean;
	daysOverdue: number;
}

function getSessionKey(token: string) {
	return `overdue-modal-shown-${token}`;
}

export function OverdueBillsModal() {
	const { data: session } = useSession();
	const [open, setOpen] = useState(false);
	const [bills, setBills] = useState<OverdueBill[]>([]);

	useEffect(() => {
		const token = session?.session?.token;
		if (!token) return;

		// Show once per auth session (cleared on logout/new login)
		const key = getSessionKey(token);
		if (sessionStorage.getItem(key)) return;

		fetch("/api/bills/overdue")
			.then((r) => r.json())
			.then((data: OverdueBill[]) => {
				if (Array.isArray(data) && data.length > 0) {
					setBills(data);
					setOpen(true);
				}
			})
			.catch(() => {});
	}, [session?.session?.token]);

	function handleDismiss() {
		const token = session?.session?.token;
		if (token) {
			sessionStorage.setItem(getSessionKey(token), "1");
		}
		setOpen(false);
	}

	const overdue = bills.filter((b) => b.isOverdue);
	const upcoming = bills.filter((b) => !b.isOverdue);

	return (
		<Dialog
			open={open}
			onOpenChange={(v) => {
				if (!v) handleDismiss();
			}}
		>
			<DialogContent className="max-w-md">
				<DialogHeader>
					<DialogTitle className="flex items-center gap-2 text-base">
						<AlertTriangle className="h-5 w-5 text-amber-400 shrink-0" />
						Atenção: contas a vencer
					</DialogTitle>
				</DialogHeader>

				<div className="space-y-4">
					{overdue.length > 0 && (
						<div>
							<p className="text-xs font-semibold text-red-400 uppercase tracking-wide mb-2">
								Vencidas
							</p>
							<ul className="space-y-2">
								{overdue.map((bill) => (
									<li
										key={bill.id}
										className="flex items-center justify-between rounded-lg border border-red-900/50 bg-red-950/30 px-3 py-2.5"
									>
										<div>
											<p className="text-sm font-medium text-zinc-100">
												{bill.name}
											</p>
											<p className="text-xs text-red-400">
												Dia {bill.dueDay} — {bill.daysOverdue} dia(s) em atraso
											</p>
										</div>
										<span className="text-sm font-semibold text-red-400 ml-3 shrink-0">
											{formatCurrency(bill.amount)}
										</span>
									</li>
								))}
							</ul>
						</div>
					)}

					{upcoming.length > 0 && (
						<div>
							<p className="text-xs font-semibold text-amber-400 uppercase tracking-wide mb-2">
								Próximas do vencimento
							</p>
							<ul className="space-y-2">
								{upcoming.map((bill) => (
									<li
										key={bill.id}
										className="flex items-center justify-between rounded-lg border border-amber-900/50 bg-amber-950/20 px-3 py-2.5"
									>
										<div>
											<p className="text-sm font-medium text-zinc-100">
												{bill.name}
											</p>
											<p className="text-xs text-amber-400 flex items-center gap-1">
												<Clock className="h-3 w-3" />
												{bill.daysUntil === 0
													? "Vence hoje"
													: `Vence em ${bill.daysUntil} dia(s)`}
											</p>
										</div>
										<span className="text-sm font-semibold text-amber-400 ml-3 shrink-0">
											{formatCurrency(bill.amount)}
										</span>
									</li>
								))}
							</ul>
						</div>
					)}

					<div className="flex gap-2 pt-1">
						<Button asChild className="flex-1">
							<Link href="/contas" onClick={handleDismiss}>
								Ver contas fixas
							</Link>
						</Button>
						<Button
							variant="outline"
							className="flex-1"
							onClick={handleDismiss}
						>
							Ignorar
						</Button>
					</div>
				</div>
			</DialogContent>
		</Dialog>
	);
}
