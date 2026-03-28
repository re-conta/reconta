import { Header } from "@/components/layout/header";
import { CategoriasClient } from "@/components/categories/categorias-client";

export default function SharedCategoriasPage() {
	return (
		<>
			<Header
				title="Categorias"
				description="Visualizando categorias compartilhadas"
			/>
			<CategoriasClient />
		</>
	);
}
