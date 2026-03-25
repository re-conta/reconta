import { Header } from "@/components/layout/header";
import { TransacoesClient } from "@/components/transactions/transacoes-client";

export default function TransacoesPage() {
	return (
		<>
			<Header
				title="Lançamentos"
				description="Livro caixa — receitas e despesas"
			/>
			<TransacoesClient />
		</>
	);
}
