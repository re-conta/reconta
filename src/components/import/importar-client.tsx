"use client";

import { CheckCircle2, FileText, Upload, XCircle } from "lucide-react";
import { useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { formatCurrency, formatDate } from "@/lib/utils";
import { useAccounts } from "@/hooks/use-accounts";

interface ImportResult {
	imported: number;
	transactions: Array<{
		id: number;
		date: string;
		description: string;
		amount: number;
		type: string;
	}>;
}

export function ImportarClient() {
	const [file, setFile] = useState<File | null>(null);
	const [accountId, setAccountId] = useState("");
	const [uploading, setUploading] = useState(false);
	const [result, setResult] = useState<ImportResult | null>(null);
	const [error, setError] = useState<string | null>(null);
	const inputRef = useRef<HTMLInputElement>(null);
	const { accounts } = useAccounts();

	function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
		const f = e.target.files?.[0];
		if (f) {
			setFile(f);
			setResult(null);
			setError(null);
		}
	}

	function handleDrop(e: React.DragEvent) {
		e.preventDefault();
		const f = e.dataTransfer.files[0];
		if (f?.name.endsWith(".pdf")) {
			setFile(f);
			setResult(null);
			setError(null);
		}
	}

	async function handleUpload() {
		if (!file) return;
		setUploading(true);
		setError(null);
		setResult(null);

		const formData = new FormData();
		formData.append("file", file);
		if (accountId) formData.append("accountId", accountId);

		const res = await fetch("/api/import", { method: "POST", body: formData });
		const data = await res.json();

		if (!res.ok) {
			setError(data.error ?? "Erro ao processar o arquivo.");
		} else {
			setResult(data);
		}
		setUploading(false);
	}

	return (
		<div className="max-w-2xl space-y-6">
			{/* Upload area */}
			<Card>
				<CardHeader>
					<CardTitle>Selecionar arquivo PDF</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					{/* Drop zone */}
					<button
						onDrop={handleDrop}
						onDragOver={(e) => e.preventDefault()}
						onClick={() => inputRef.current?.click()}
						className={`border-2 border-dashed rounded-xl p-10 text-center cursor-pointer transition-colors ${
							file
								? "border-indigo-500 bg-indigo-900/10"
								: "border-zinc-700 hover:border-zinc-500"
						}`}
						type="button"
					>
						<input
							ref={inputRef}
							type="file"
							accept=".pdf"
							className="hidden"
							onChange={handleFileChange}
						/>
						{file ? (
							<div className="flex flex-col items-center gap-2">
								<FileText className="h-10 w-10 text-indigo-400" />
								<p className="text-zinc-200 font-medium">{file.name}</p>
								<p className="text-xs text-zinc-500">
									{(file.size / 1024).toFixed(1)} KB
								</p>
							</div>
						) : (
							<div className="flex flex-col items-center gap-2">
								<Upload className="h-10 w-10 text-zinc-500" />
								<p className="text-zinc-400">
									Arraste um PDF aqui ou clique para selecionar
								</p>
								<p className="text-xs text-zinc-600">
									Extratos bancários em PDF (Itaú, Bradesco, BB, Nubank, etc.)
								</p>
							</div>
						)}
					</button>

					{/* Account select */}
					<div>
						<Label>Conta bancária (opcional)</Label>
						<Select value={accountId} onValueChange={setAccountId}>
							<SelectTrigger className="mt-1">
								<SelectValue placeholder="Selecionar conta..." />
							</SelectTrigger>
							<SelectContent>
								{accounts.map((a) => (
									<SelectItem key={a.id} value={String(a.id)}>
										{a.name}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>

					<Button
						onClick={handleUpload}
						disabled={!file || uploading}
						className="w-full"
					>
						{uploading ? "Processando..." : "Importar transações"}
					</Button>
				</CardContent>
			</Card>

			{/* Error */}
			{error && (
				<div className="flex items-start gap-3 rounded-xl border border-red-800 bg-red-900/20 p-4">
					<XCircle className="h-5 w-5 text-red-400 shrink-0 mt-0.5" />
					<div>
						<p className="font-medium text-red-400">Erro ao importar</p>
						<p className="text-sm text-red-300 mt-1">{error}</p>
					</div>
				</div>
			)}

			{/* Result */}
			{result && (
				<div className="space-y-4">
					<div className="flex items-center gap-3 rounded-xl border border-emerald-800 bg-emerald-900/20 p-4">
						<CheckCircle2 className="h-5 w-5 text-emerald-400 shrink-0" />
						<p className="font-medium text-emerald-400">
							{result.imported} transação(ões) importada(s) com sucesso!
						</p>
					</div>

					<Card>
						<CardHeader>
							<CardTitle>Transações importadas</CardTitle>
						</CardHeader>
						<CardContent className="p-0">
							<div className="overflow-x-auto">
								<table className="w-full text-sm">
									<thead>
										<tr className="border-b border-zinc-800">
											<th className="text-left px-4 py-3 text-xs font-medium text-zinc-400">
												Data
											</th>
											<th className="text-left px-4 py-3 text-xs font-medium text-zinc-400">
												Descrição
											</th>
											<th className="text-right px-4 py-3 text-xs font-medium text-zinc-400">
												Valor
											</th>
										</tr>
									</thead>
									<tbody>
										{result.transactions.map((tx) => (
											<tr key={tx.id} className="border-b border-zinc-800/50">
												<td className="px-4 py-2.5 text-zinc-400 whitespace-nowrap">
													{formatDate(tx.date)}
												</td>
												<td className="px-4 py-2.5 text-zinc-200 truncate max-w-xs">
													{tx.description}
												</td>
												<td
													className={`px-4 py-2.5 text-right font-medium whitespace-nowrap ${tx.type === "income" ? "text-emerald-400" : "text-red-400"}`}
												>
													{tx.type === "income" ? "+" : "-"}
													{formatCurrency(tx.amount)}
												</td>
											</tr>
										))}
									</tbody>
								</table>
							</div>
						</CardContent>
					</Card>
				</div>
			)}

			{/* Instructions */}
			<Card>
				<CardHeader>
					<CardTitle className="text-base">Como usar</CardTitle>
				</CardHeader>
				<CardContent className="space-y-2 text-sm text-zinc-400">
					<p>
						1. Acesse o site do seu banco e faça o download do extrato em PDF.
					</p>
					<p>
						2. Arraste o arquivo para a área acima ou clique para selecioná-lo.
					</p>
					<p>3. Selecione a conta bancária correspondente (opcional).</p>
					<p>4. Clique em "Importar transações".</p>
					<p className="text-zinc-500 text-xs mt-3">
						O sistema tenta detectar automaticamente o formato do extrato. Caso
						as transações não sejam reconhecidas, verifique se o PDF contém
						texto selecionável (não apenas imagem).
					</p>
				</CardContent>
			</Card>
		</div>
	);
}
