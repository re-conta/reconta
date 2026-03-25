"use client";

import {
	AlertCircle,
	ArrowDownRight,
	ArrowUpRight,
	ChevronLeft,
	ChevronRight,
	TrendingDown,
	TrendingUp,
	Wallet,
} from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
//import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatCurrency, formatDate, formatMonth } from "@/lib/utils";
import { MonthlyBalanceChart } from "./monthly-balance-chart";
import { SpendingPieChart } from "./spending-pie-chart";

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
	pendingBills: Array<{
		id: number;
		name: string;
		amount: number;
		dueDay: number;
		categoryName: string | null;
	}>;
	monthlyBalance: Array<{
		month: number;
		year: number;
		income: number;
		expense: number;
		balance: number;
	}>;
}

interface Props {
	initialMonth: number;
	initialYear: number;
}

function pct(curr: number, prev: number) {
	if (prev === 0) return null;
	return ((curr - prev) / prev) * 100;
}

export function DashboardClient({ initialMonth, initialYear }: Props) {
	const [month, setMonth] = useState(initialMonth);
	const [year, setYear] = useState(initialYear);
	const [data, setData] = useState<DashboardData | null>(null);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		setLoading(true);
		fetch(`/api/dashboard?month=${month}&year=${year}`)
			.then((r) => {
				if (!r.ok) throw new Error(r.statusText);
				return r.json();
			})
			.then((d) => {
				setData(d);
				setLoading(false);
			})
			.catch(() => setLoading(false));
	}, [month, year]);

	function prevMonth() {
		if (month === 1) {
			setMonth(12);
			setYear((y) => y - 1);
		} else setMonth((m) => m - 1);
	}

	function nextMonth() {
		if (month === 12) {
			setMonth(1);
			setYear((y) => y + 1);
		} else setMonth((m) => m + 1);
	}

	const today = new Date();
	const isCurrentMonth =
		month === today.getMonth() + 1 && year === today.getFullYear();

	return (
		<div className="space-y-6">
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
							previous={data.previous.income}
							icon={<ArrowUpRight className="h-5 w-5 text-emerald-400" />}
							color="emerald"
						/>
						<KpiCard
							title="Despesas"
							value={data.current.expense}
							previous={data.previous.expense}
							icon={<ArrowDownRight className="h-5 w-5 text-red-400" />}
							color="red"
							invertTrend
						/>
						<KpiCard
							title="Saldo"
							value={data.current.balance}
							previous={data.previous.balance}
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

					{/* Pending bills + Recent transactions */}
					<div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
						{/* Pending bills */}
						<Card>
							<CardHeader>
								<div className="flex items-center justify-between">
									<CardTitle className="flex items-center gap-2">
										<AlertCircle className="h-4 w-4 text-amber-400" />
										Contas pendentes
									</CardTitle>
									<Link
										href="/contas"
										className="text-xs text-indigo-400 hover:underline"
									>
										Ver todas
									</Link>
								</div>
							</CardHeader>
							<CardContent>
								{data.pendingBills.length === 0 ? (
									<p className="text-sm text-zinc-500 text-center py-4">
										Todas as contas estão pagas! 🎉
									</p>
								) : (
									<ul className="space-y-2">
										{data.pendingBills.slice(0, 5).map((bill) => {
											const today = new Date();
											const dueDate = new Date(
												today.getFullYear(),
												today.getMonth(),
												bill.dueDay,
											);
											const isOverdue = dueDate < today;
											return (
												<li
													key={bill.id}
													className="flex items-center justify-between py-2 border-b border-zinc-800 last:border-0"
												>
													<div>
														<p className="text-sm font-medium text-zinc-200">
															{bill.name}
														</p>
														<p
															className={`text-xs ${isOverdue ? "text-red-400" : "text-zinc-500"}`}
														>
															Vence dia {bill.dueDay}
															{isOverdue ? " · Atrasada!" : ""}
														</p>
													</div>
													<span className="text-sm font-medium text-amber-400">
														{formatCurrency(bill.amount)}
													</span>
												</li>
											);
										})}
									</ul>
								)}
							</CardContent>
						</Card>

						{/* Recent transactions */}
						<Card>
							<CardHeader>
								<div className="flex items-center justify-between">
									<CardTitle>Últimos lançamentos</CardTitle>
									<Link
										href="/transacoes"
										className="text-xs text-indigo-400 hover:underline"
									>
										Ver todos
									</Link>
								</div>
							</CardHeader>
							<CardContent>
								{data.recentTransactions.length === 0 ? (
									<p className="text-sm text-zinc-500 text-center py-4">
										Nenhum lançamento neste mês.
									</p>
								) : (
									<ul className="space-y-2">
										{data.recentTransactions.map((tx) => (
											<li
												key={tx.id}
												className="flex items-center justify-between py-2 border-b border-zinc-800 last:border-0"
											>
												<div className="min-w-0">
													<p className="text-sm font-medium text-zinc-200 truncate">
														{tx.description}
													</p>
													<p className="text-xs text-zinc-500">
														{formatDate(tx.date)}
														{tx.categoryName ? ` · ${tx.categoryName}` : ""}
													</p>
												</div>
												<span
													className={`text-sm font-semibold ml-4 shrink-0 ${tx.type === "income" ? "text-emerald-400" : "text-red-400"}`}
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
					</div>
				</>
			) : null}
		</div>
	);
}

function KpiCard({
	title,
	value,
	previous,
	icon,
	invertTrend = false,
}: {
	title: string;
	value: number;
	previous: number;
	icon: React.ReactNode;
	color: "emerald" | "red" | "indigo";
	invertTrend?: boolean;
}) {
	const change = pct(value, previous);
	const isPositive = change !== null && (invertTrend ? change < 0 : change > 0);
	const isNegative = change !== null && (invertTrend ? change > 0 : change < 0);

	return (
		<Card>
			<CardContent className="pt-5">
				<div className="flex items-center justify-between mb-3">
					<span className="text-sm text-zinc-400">{title}</span>
					{icon}
				</div>
				<p className="text-2xl font-bold text-zinc-100">
					{formatCurrency(value)}
				</p>
				{change !== null && (
					<div className="flex items-center gap-1 mt-2">
						{isPositive ? (
							<TrendingUp className="h-3.5 w-3.5 text-emerald-400" />
						) : (
							<TrendingDown className="h-3.5 w-3.5 text-red-400" />
						)}
						<span
							className={`text-xs ${isPositive ? "text-emerald-400" : isNegative ? "text-red-400" : "text-zinc-500"}`}
						>
							{change > 0 ? "+" : ""}
							{change.toFixed(1)}% vs mês anterior
						</span>
					</div>
				)}
				{change === null && (
					<p className="text-xs text-zinc-600 mt-2">Sem dados anteriores</p>
				)}
			</CardContent>
		</Card>
	);
}
