"use client";

import {
	ArrowDownRight,
	ArrowLeft,
	ArrowUpRight,
	ChevronLeft,
	ChevronRight,
	Wallet,
} from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatCurrency, formatDate, formatMonth } from "@/lib/utils";
import { MonthlyBalanceChart } from "@/components/dashboard/monthly-balance-chart";
import { SpendingPieChart } from "@/components/dashboard/spending-pie-chart";

interface DashboardData {
	current: { income: number; expense: number; balance: number };
	previous: { income: number; expense: number; balance: number };
	expensesByCategory: Array<{
		categoryName: string | null;
		categoryColor: string | null;
		total: number;
	}>;
	recentTransactions: Array<{
		id: number;
		date: string;
		description: string;
		amount: number;
		type: string;
		categoryName: string | null;
		categoryColor: string | null;
	}>;
	monthlyBalance: Array<{
		month: number;
		year: number;
		income: number;
		expense: number;
		balance: number;
	}>;
}

interface OwnerInfo {
	ownerName: string;
	ownerEmail: string;
	shares: Array<{
		id: number;
		scope: "all" | "yearly" | "monthly";
		scopeMonth: number | null;
		scopeYear: number | null;
	}>;
}

interface KpiCardProps {
	title: string;
	value: number;
	icon: React.ReactNode;
	color: "emerald" | "red" | "indigo";
}

function KpiCard({ title, value, icon, color }: KpiCardProps) {
	const colors = {
		emerald: "text-emerald-400",
		red: "text-red-400",
		indigo: "text-indigo-400",
	};
	return (
		<Card>
			<CardContent className="p-5">
				<div className="flex items-center justify-between mb-3">
					<p className="text-sm text-zinc-400">{title}</p>
					{icon}
				</div>
				<p className={`text-2xl font-bold ${colors[color]}`}>
					{formatCurrency(value)}
				</p>
			</CardContent>
		</Card>
	);
}

interface SharedViewClientProps {
	ownerId: string;
}

export function SharedViewClient({ ownerId }: SharedViewClientProps) {
	const today = new Date();
	const [month, setMonth] = useState(today.getMonth() + 1);
	const [year, setYear] = useState(today.getFullYear());
	const [ownerInfo, setOwnerInfo] = useState<OwnerInfo | null>(null);
	const [data, setData] = useState<DashboardData | null>(null);
	const [loading, setLoading] = useState(true);
	const [forbidden, setForbidden] = useState(false);

	useEffect(() => {
		fetch(`/api/shared/${ownerId}/info`)
			.then((r) => {
				if (r.status === 403) {
					setForbidden(true);
					return null;
				}
				return r.json();
			})
			.then((d) => {
				if (d) setOwnerInfo(d);
			});
	}, [ownerId]);

	useEffect(() => {
		setLoading(true);
		fetch(`/api/shared/${ownerId}/dashboard?month=${month}&year=${year}`)
			.then((r) => {
				if (r.status === 403) {
					setForbidden(true);
					return null;
				}
				return r.json();
			})
			.then((d) => {
				if (d) setData(d);
				setLoading(false);
			})
			.catch(() => setLoading(false));
	}, [ownerId, month, year]);

	function prevMonth() {
		if (month === 1) {
			setMonth(12);
			setYear((y) => y - 1);
		} else {
			setMonth((m) => m - 1);
		}
	}

	function nextMonth() {
		const now = new Date();
		const isCurrentMonth =
			month === now.getMonth() + 1 && year === now.getFullYear();
		if (isCurrentMonth) return;
		if (month === 12) {
			setMonth(1);
			setYear((y) => y + 1);
		} else {
			setMonth((m) => m + 1);
		}
	}

	const isCurrentMonth =
		month === today.getMonth() + 1 && year === today.getFullYear();

	if (forbidden) {
		return (
			<div className="flex flex-col items-center justify-center h-64 gap-4">
				<p className="text-zinc-400">Você não tem acesso a esses dados.</p>
				<Link href="/compartilhamentos">
					<Button variant="outline" size="sm">
						<ArrowLeft className="h-4 w-4 mr-2" />
						Voltar
					</Button>
				</Link>
			</div>
		);
	}

	return (
		<div className="space-y-6">
			{/* Header */}
			<div className="flex items-start gap-4">
				<Link href="/compartilhamentos">
					<Button variant="ghost" size="icon" className="mt-0.5">
						<ArrowLeft className="h-4 w-4" />
					</Button>
				</Link>
				<div>
					<h1 className="text-2xl font-semibold text-zinc-100">
						{ownerInfo ? ownerInfo.ownerName : "Carregando..."}
					</h1>
					{ownerInfo && (
						<p className="text-sm text-zinc-400 mt-0.5">
							Visualizando dados financeiros compartilhados
						</p>
					)}
				</div>
			</div>

			{/* Month selector */}
			<div className="flex items-center gap-3">
				<Button variant="outline" size="icon" onClick={prevMonth}>
					<ChevronLeft className="h-4 w-4" />
				</Button>
				<span className="text-zinc-100 font-medium capitalize min-w-40 text-center">
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

			{loading ? (
				<div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
					{[0, 1, 2].map((i) => (
						<Card key={i} className="animate-pulse">
							<CardContent className="h-28" />
						</Card>
					))}
				</div>
			) : data ? (
				<>
					{/* KPI cards */}
					<div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
						<KpiCard
							title="Receitas"
							value={data.current.income}
							icon={<ArrowUpRight className="h-5 w-5 text-emerald-400" />}
							color="emerald"
						/>
						<KpiCard
							title="Despesas"
							value={data.current.expense}
							icon={<ArrowDownRight className="h-5 w-5 text-red-400" />}
							color="red"
						/>
						<KpiCard
							title="Saldo"
							value={data.current.balance}
							icon={<Wallet className="h-5 w-5 text-indigo-400" />}
							color="indigo"
						/>
					</div>

					{/* Charts */}
					<div className="grid grid-cols-1 lg:grid-cols-5 gap-4">
						<Card className="lg:col-span-3">
							<CardHeader>
								<CardTitle>Evolução dos últimos 6 meses</CardTitle>
							</CardHeader>
							<CardContent>
								<MonthlyBalanceChart data={data.monthlyBalance} />
							</CardContent>
						</Card>

						<Card className="lg:col-span-2">
							<CardHeader>
								<CardTitle>Gastos por categoria</CardTitle>
							</CardHeader>
							<CardContent>
								<SpendingPieChart data={data.expensesByCategory} />
							</CardContent>
						</Card>
					</div>

					{/* Recent transactions */}
					<Card>
						<CardHeader>
							<CardTitle>Últimos lançamentos</CardTitle>
						</CardHeader>
						<CardContent>
							{data.recentTransactions.length === 0 ? (
								<p className="text-sm text-zinc-500 text-center py-4">
									Nenhum lançamento neste mês.
								</p>
							) : (
								<ul className="divide-y divide-zinc-800">
									{data.recentTransactions.map((tx) => (
										<li
											key={tx.id}
											className="flex items-center justify-between py-2"
										>
											<div className="min-w-0 flex-1">
												<p className="text-sm font-medium text-zinc-200 truncate">
													{tx.description}
												</p>
												<p className="text-xs text-zinc-500">
													{formatDate(tx.date)}
													{tx.categoryName && ` · ${tx.categoryName}`}
												</p>
											</div>
											<span
												className={`text-sm font-medium ml-4 shrink-0 ${
													tx.type === "income"
														? "text-emerald-400"
														: "text-red-400"
												}`}
											>
												{tx.type === "income" ? "+" : "-"}
												{formatCurrency(tx.amount)}
											</span>
										</li>
									))}
								</ul>
							)}
						</CardContent>
					</Card>
				</>
			) : (
				<p className="text-zinc-400 text-center py-8">
					Nenhum dado encontrado para este período.
				</p>
			)}
		</div>
	);
}
