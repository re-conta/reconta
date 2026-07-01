import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export function formatCurrency(value: number): string {
	return new Intl.NumberFormat("pt-BR", {
		style: "currency",
		currency: "BRL",
	}).format(value);
}

export function formatDate(date: string | Date): string {
	const d = typeof date === "string" ? new Date(`${date}T00:00:00`) : date;
	return new Intl.DateTimeFormat("pt-BR").format(d);
}

export function formatMonth(month: number, year: number): string {
	const date = new Date(year, month - 1, 1);
	return new Intl.DateTimeFormat("pt-BR", {
		month: "long",
		year: "numeric",
	}).format(date);
}

export function getCurrentMonth(): { month: number; year: number } {
	const now = new Date();
	return { month: now.getMonth() + 1, year: now.getFullYear() };
}

export function getMonthRange(
	month: number,
	year: number,
): { start: string; end: string } {
	const start = new Date(year, month - 1, 1);
	const end = new Date(year, month, 0);
	return {
		start: start.toISOString().split("T")[0],
		end: end.toISOString().split("T")[0],
	};
}

export function getPreviousMonth(
	month: number,
	year: number,
): { month: number; year: number } {
	if (month === 1) return { month: 12, year: year - 1 };
	return { month: month - 1, year };
}

export function getYearRange(year: number): { start: string; end: string } {
	return { start: `${year}-01-01`, end: `${year}-12-31` };
}

export function getPreviousPeriod(
	start: string,
	end: string,
): { start: string; end: string } {
	const startDate = new Date(`${start}T00:00:00`);
	const endDate = new Date(`${end}T00:00:00`);
	const days =
		Math.round((endDate.getTime() - startDate.getTime()) / 86400000) + 1;

	const prevEnd = new Date(startDate);
	prevEnd.setDate(prevEnd.getDate() - 1);
	const prevStart = new Date(prevEnd);
	prevStart.setDate(prevStart.getDate() - days + 1);

	return {
		start: prevStart.toISOString().split("T")[0],
		end: prevEnd.toISOString().split("T")[0],
	};
}

export const CATEGORY_ICONS: Record<string, string> = {
	utensils: "🍽️",
	home: "🏠",
	car: "🚗",
	heart: "❤️",
	book: "📚",
	smile: "😊",
	shirt: "👕",
	laptop: "💻",
	zap: "⚡",
	"more-horizontal": "•••",
	briefcase: "💼",
	"trending-up": "📈",
	"plus-circle": "➕",
	circle: "⚪",
};
