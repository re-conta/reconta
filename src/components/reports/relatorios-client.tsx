"use client";

import {
	ChevronLeft,
	ChevronRight,
	TrendingDown,
	TrendingUp,
} from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatMonth } from "@/lib/utils";
import { MonthlyBalanceChart } from "@/components/dashboard/monthly-balance-chart";
import { SpendingPieChart } from "@/components/dashboard/spending-pie-chart";
import { SavingsChart } from "./savings-chart";
import { ComparisonChart } from "./comparison-chart";

interface DashboardData {
	current: { income: number; expense: number; balance: number };
	previous: { income: number; expense: number; balance: number };
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

export function RelatoriosClient({
	initialMonth,
	initialYear,
}: {
	initialMonth: number;
	initialYear: number;
}) {
	const [month, setMonth] = useState(initialMonth);
	const [year, setYear] = useState(initialYear);
	const [data, setData] = useState<DashboardData | null>(null);
	const [loading, setLoading] = useState(true);

	const today = new Date();
	const isCurrentMonth =
		month === today.getMonth() + 1 && year === today.getFullYear();

	useEffect(() => {
		setLoading(true);
		fetch(`/api/dashboard?month=${month}&year=${year}`)
			.then((r) => r.json())
			.then((d) => {
				setData(d);
				setLoading(false);
			});
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

	return (
		<div className="space-y-6">
			{/* Month selector */}
			<div className="flex items-center gap-2">
				<Button variant="outline" size="icon" onClick={prevMonth}>
					<ChevronLeft className="h-4 w-4" />
				</Button>
				<span className="font-medium capitalize min-w-[160px] text-center text-zinc-100">
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

			{loading || !data ? (
				<div className="text-center text-zinc-500 py-12 text-sm">
					Carregando dados...
				</div>
			) : (
				<>
					{/* Month comparison */}
					<Card>
						<CardHeader>
							<CardTitle>Comparativo com mês anterior</CardTitle>
						</CardHeader>
						<CardContent>
							<ComparisonChart
								current={data.current}
								previous={data.previous}
							/>
						</CardContent>
					</Card>

					{/* Savings / Guardando dinheiro */}
					<Card>
						<CardHeader>
							<div className="flex items-center justify-between">
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
