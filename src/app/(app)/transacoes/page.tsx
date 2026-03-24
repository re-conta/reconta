import { Header } from "@/components/layout/header";
import { TransacoesClient } from "@/components/transactions/transacoes-client";
import { getCurrentMonth } from "@/lib/utils";

export default function TransacoesPage() {
	const { month, year } = getCurrentMonth();
	return (
		<>
			<Header
				title="Lançamentos"
				description="Livro caixa — receitas e despesas"
			/>
			<TransacoesClient initialMonth={month} initialYear={year} />
		</>
	);
}
