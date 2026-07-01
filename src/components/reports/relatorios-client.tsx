"use client";

import {
	ChevronLeft,
	ChevronRight,
	TrendingDown,
	TrendingUp,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatCurrency, formatMonth, getCurrentMonth } from "@/lib/utils";
import { MonthlyBalanceChart } from "@/components/dashboard/monthly-balance-chart";
import { SpendingPieChart } from "@/components/dashboard/spending-pie-chart";
import { useSharedOwner } from "@/components/layout/shared-owner-context";
import { SavingsChart } from "./savings-chart";
import { ComparisonChart } from "./comparison-chart";

type Scope = "month" | "year" | "all" | "custom";

interface DashboardData {
	current: { income: number; expense: number; balance: number };
	previous: { income: number; expense: number; balance: number };
	hasComparison: boolean;
	expensesByCategory: Array<{
		categoryName: string | null;
		categoryColor: string | null;
		total: number;
	}>;
	monthlyBalance: Array<{
		month: number;
		year: number;
		income: number;
		expense: number;
		balance: number;
	}>;
}

const SCOPE_LABELS: Record<Scope, string> = {
	month: "Mês",
	year: "Ano",
	all: "Todo o período",
	custom: "Período personalizado",
};

function todayIso() {
	return new Date().toISOString().split("T")[0];
}

export function RelatoriosClient() {
	const current = getCurrentMonth();
	const [scope, setScope] = useState<Scope>("month");
	const [month, setMonth] = useState(current.month);
	const [year, setYear] = useState(current.year);
	const [customStart, setCustomStart] = useState(todayIso());
	const [customEnd, setCustomEnd] = useState(todayIso());

	const [data, setData] = useState<DashboardData | null>(null);
	const [loading, setLoading] = useState(true);
	const shared = useSharedOwner();
	const apiBase = shared ? shared.apiBase : "/api";

	const today = new Date();
	const isCurrentMonth =
		month === today.getMonth() + 1 && year === today.getFullYear();
	const isCurrentYear = year === today.getFullYear();

	const query = useMemo(() => {
		const params = new URLSearchParams({ scope });
		if (scope === "month") {
			params.set("month", String(month));
			params.set("year", String(year));
		} else if (scope === "year") {
			params.set("year", String(year));
		} else if (scope === "custom") {
			params.set("start", customStart);
			params.set("end", customEnd);
		}
		return params.toString();
	}, [scope, month, year, customStart, customEnd]);

	useEffect(() => {
		setLoading(true);
		fetch(`${apiBase}/dashboard?${query}`)
			.then((r) => {
				if (!r.ok) throw new Error(r.statusText);
				return r.json();
			})
			.then((d) => {
				setData(d);
				setLoading(false);
			})
			.catch(() => setLoading(false));
	}, [apiBase, query]);

	function prevMonth() {
		if (month === 1) {
			setMonth(12);
			setYear((y) => y - 1);
		} else {
			setMonth((m) => m - 1);
		}
	}
	function nextMonth() {
		if (month === 12) {
			setMonth(1);
			setYear((y) => y + 1);
		} else {
			setMonth((m) => m + 1);
		}
	}

	const periodLabel =
		scope === "month"
			? formatMonth(month, year)
			: scope === "year"
				? String(year)
				: scope === "all"
					? "Todo o período"
					: "Período personalizado";

	return (
		<div className="space-y-6">
			{/* Period selector */}
			<Card>
				<CardContent className="flex flex-col gap-3 pt-6 sm:flex-row sm:items-center sm:justify-between">
					<div className="flex flex-wrap items-center gap-2">
						<Select
							value={scope}
							onValueChange={(v) => setScope(v as Scope)}
						>
							<SelectTrigger className="w-44">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								{Object.entries(SCOPE_LABELS).map(([value, label]) => (
									<SelectItem key={value} value={value}>
										{label}
									</SelectItem>
								))}
							</SelectContent>
						</Select>

						{scope === "month" && (
							<div className="flex items-center gap-2">
								<Button variant="outline" size="icon" onClick={prevMonth}>
									<ChevronLeft className="h-4 w-4" />
								</Button>
								<span className="font-medium capitalize min-w-36 text-center text-zinc-100">
									{periodLabel}
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
						)}

						{scope === "year" && (
							<div className="flex items-center gap-2">
								<Button
									variant="outline"
									size="icon"
									onClick={() => setYear((y) => y - 1)}
								>
									<ChevronLeft className="h-4 w-4" />
								</Button>
								<span className="font-medium min-w-20 text-center text-zinc-100">
									{periodLabel}
								</span>
								<Button
									variant="outline"
									size="icon"
									onClick={() => setYear((y) => y + 1)}
									disabled={isCurrentYear}
								>
									<ChevronRight className="h-4 w-4" />
								</Button>
							</div>
						)}

						{scope === "custom" && (
							<div className="flex flex-wrap items-center gap-2">
								<Input
									type="date"
									value={customStart}
									max={customEnd}
									onChange={(e) => setCustomStart(e.target.value)}
									className="w-auto"
								/>
								<span className="text-zinc-500 text-sm">até</span>
								<Input
									type="date"
									value={customEnd}
									min={customStart}
									max={todayIso()}
									onChange={(e) => setCustomEnd(e.target.value)}
									className="w-auto"
								/>
							</div>
						)}
					</div>
				</CardContent>
			</Card>

			{loading || !data ? (
				<div className="text-center text-zinc-500 py-12 text-sm">
					Carregando dados...
				</div>
			) : (
				<>
					{/* Period comparison */}
					{data.hasComparison && (
						<Card>
							<CardHeader>
								<CardTitle>
									{scope === "year"
										? "Comparativo com ano anterior"
										: scope === "custom"
											? "Comparativo com período anterior equivalente"
											: "Comparativo com mês anterior"}
								</CardTitle>
							</CardHeader>
							<CardContent>
								<ComparisonChart
									current={data.current}
									previous={data.previous}
								/>
							</CardContent>
						</Card>
					)}

					{!data.hasComparison && (
						<Card>
							<CardHeader>
								<CardTitle>Resumo — {periodLabel}</CardTitle>
							</CardHeader>
							<CardContent>
								<div className="grid grid-cols-3 gap-3">
									<SummaryStat label="Receitas" value={data.current.income} />
									<SummaryStat label="Despesas" value={data.current.expense} />
									<SummaryStat label="Saldo" value={data.current.balance} />
								</div>
							</CardContent>
						</Card>
					)}

					{/* Savings / Guardando dinheiro */}
					<Card>
						<CardHeader>
							<div className="flex flex-wrap items-center justify-between gap-2">
								<CardTitle>Capacidade de poupança — últimos 6 meses</CardTitle>
								<SavingsIndicator
									balance={data.current.balance}
									income={data.current.income}
								/>
							</div>
						</CardHeader>
						<CardContent>
							<SavingsChart data={data.monthlyBalance} />
						</CardContent>
					</Card>

					<div className="grid grid-cols-1 lg:grid-cols-5 gap-4">
						<Card className="lg:col-span-3">
							<CardHeader>
								<CardTitle>Receitas × Despesas × Saldo</CardTitle>
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
				</>
			)}
		</div>
	);
}

function SummaryStat({ label, value }: { label: string; value: number }) {
	return (
		<div className="rounded-lg bg-zinc-800/50 p-3 text-center">
			<p className="text-xs text-zinc-500 mb-1">{label}</p>
			<p className="text-sm font-bold text-zinc-100">
				{formatCurrency(value)}
			</p>
		</div>
	);
}

function SavingsIndicator({
	balance,
	income,
}: {
	balance: number;
	income: number;
}) {
	if (income === 0) return null;
	const rate = (balance / income) * 100;
	const isPositive = balance >= 0;

	return (
		<div
			className={`flex items-center gap-2 text-sm font-medium ${isPositive ? "text-emerald-400" : "text-red-400"}`}
		>
			{isPositive ? (
				<TrendingUp className="h-4 w-4" />
			) : (
				<TrendingDown className="h-4 w-4" />
			)}
			{isPositive ? "+" : ""}
			{rate.toFixed(1)}% de poupança
		</div>
	);
}
