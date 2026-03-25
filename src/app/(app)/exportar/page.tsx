import { Header } from "@/components/layout/header";
import { ExportarClient } from "@/components/export/exportar-client";

export default function ExportarPage() {
	return (
		<>
			<Header
				title="Exportar"
				description="Baixe seus lançamentos em planilha ou PDF"
			/>
			<ExportarClient />
		</>
	);
}
