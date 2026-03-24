import { eq } from "drizzle-orm";
import { db } from "./index";
import { accounts, categories } from "./schema";

export async function seedUserDefaults(userId: string) {
	const existing = await db
		.select()
		.from(categories)
		.where(eq(categories.userId, userId))
		.limit(1);

	if (existing.length > 0) return;

	await db.insert(categories).values([
		{
			userId,
			name: "Alimentação",
			color: "#f97316",
			icon: "utensils",
			type: "expense",
		},
		{
			userId,
			name: "Moradia",
			color: "#8b5cf6",
			icon: "home",
			type: "expense",
		},
		{
			userId,
			name: "Transporte",
			color: "#3b82f6",
			icon: "car",
			type: "expense",
		},
		{ userId, name: "Saúde", color: "#ef4444", icon: "heart", type: "expense" },
		{
			userId,
			name: "Educação",
			color: "#06b6d4",
			icon: "book",
			type: "expense",
		},
		{ userId, name: "Lazer", color: "#ec4899", icon: "smile", type: "expense" },
		{
			userId,
			name: "Vestuário",
			color: "#a855f7",
			icon: "shirt",
			type: "expense",
		},
		{
			userId,
			name: "Tecnologia",
			color: "#64748b",
			icon: "laptop",
			type: "expense",
		},
		{
			userId,
			name: "Contas & Serviços",
			color: "#f59e0b",
			icon: "zap",
			type: "expense",
		},
		{
			userId,
			name: "Outros Gastos",
			color: "#6b7280",
			icon: "more-horizontal",
			type: "expense",
		},
		{
			userId,
			name: "Salário",
			color: "#10b981",
			icon: "briefcase",
			type: "income",
		},
		{
			userId,
			name: "Freelance",
			color: "#14b8a6",
			icon: "laptop",
			type: "income",
		},
		{
			userId,
			name: "Investimentos",
			color: "#22c55e",
			icon: "trending-up",
			type: "income",
		},
		{
			userId,
			name: "Outros Ganhos",
			color: "#84cc16",
			icon: "plus-circle",
			type: "income",
		},
	]);

	await db
		.insert(accounts)
		.values([
			{ userId, name: "Conta Principal", type: "checking", balance: 0 },
		]);
}
