import { Header } from "@/components/layout/header";
import { ExportarClient } from "@/components/export/exportar-client";

export default function SharedExportarPage() {
	return (
		<>
			<Header title="Exportar" description="Exportar dados compartilhados" />
			<ExportarClient />
		</>
	);
}
