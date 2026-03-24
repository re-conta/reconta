import { Header } from "@/components/layout/header";
import { ImportarClient } from "@/components/import/importar-client";

export default function ImportarPage() {
	return (
		<>
			<Header
				title="Importar Extrato"
				description="Importe suas transações a partir de um extrato bancário em PDF"
			/>
			<ImportarClient />
		</>
	);
}
