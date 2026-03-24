import { Header } from "@/components/layout/header";
import { CategoriasClient } from "@/components/categories/categorias-client";

export default function CategoriasPage() {
	return (
		<>
			<Header
				title="Categorias"
				description="Gerencie as categorias de receitas e despesas"
			/>
			<CategoriasClient />
		</>
	);
}
