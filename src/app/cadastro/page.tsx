"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { signUp } from "@/lib/auth-client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function CadastroPage() {
	const router = useRouter();

	const [name, setName] = useState("");
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [confirm, setConfirm] = useState("");
	const [error, setError] = useState<string | null>(null);
	const [loading, setLoading] = useState(false);

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		setError(null);

		if (password !== confirm) {
			setError("As senhas não coincidem.");
			return;
		}

		if (password.length < 8) {
			setError("A senha deve ter pelo menos 8 caracteres.");
			return;
		}

		setLoading(true);

		const { error } = await signUp.email({ name, email, password });

		if (error) {
			setError(error.message ?? "Erro ao criar conta.");
			setLoading(false);
			return;
		}

		router.push("/");
		router.refresh();
	}

	return (
		<div className="min-h-screen bg-zinc-950 flex items-center justify-center p-4">
			<div className="w-full max-w-sm">
				{/* Logo */}
				<div className="flex items-center justify-center gap-2 mb-8">
					<div className="h-10 w-10 rounded-xl bg-indigo-600 flex items-center justify-center">
						<span className="text-lg font-bold text-white">R</span>
					</div>
					<div>
						<span className="text-2xl font-bold text-white">ReConta</span>
						<span className="text-sm text-zinc-400">.app</span>
					</div>
				</div>

				<div className="bg-zinc-900 border border-zinc-800 rounded-2xl p-8">
					<h1 className="text-xl font-semibold text-zinc-100 mb-1">
						Criar conta
					</h1>
					<p className="text-sm text-zinc-400 mb-6">
						Suas finanças, organizadas e privadas
					</p>

					<form onSubmit={handleSubmit} className="space-y-4">
						<div>
							<Label htmlFor="name">Nome</Label>
							<Input
								id="name"
								type="text"
								placeholder="Seu nome"
								value={name}
								onChange={(e) => setName(e.target.value)}
								required
								autoComplete="name"
								className="mt-1"
							/>
						</div>

						<div>
							<Label htmlFor="email">E-mail</Label>
							<Input
								id="email"
								type="email"
								placeholder="seu@email.com"
								value={email}
								onChange={(e) => setEmail(e.target.value)}
								required
								autoComplete="email"
								className="mt-1"
							/>
						</div>

						<div>
							<Label htmlFor="password">Senha</Label>
							<Input
								id="password"
								type="password"
								placeholder="Mínimo 8 caracteres"
								value={password}
								onChange={(e) => setPassword(e.target.value)}
								required
								autoComplete="new-password"
								className="mt-1"
							/>
						</div>

						<div>
							<Label htmlFor="confirm">Confirmar senha</Label>
							<Input
								id="confirm"
								type="password"
								placeholder="••••••••"
								value={confirm}
								onChange={(e) => setConfirm(e.target.value)}
								required
								autoComplete="new-password"
								className="mt-1"
							/>
						</div>

						{error && (
							<p className="text-sm text-red-400 bg-red-900/20 border border-red-800/30 rounded-lg px-3 py-2">
								{error}
							</p>
						)}

						<Button type="submit" className="w-full" disabled={loading}>
							{loading ? "Criando conta..." : "Criar conta"}
						</Button>
					</form>

					<p className="text-center text-sm text-zinc-500 mt-6">
						Já tem uma conta?{" "}
						<Link href="/login" className="text-indigo-400 hover:underline">
							Entrar
						</Link>
					</p>
				</div>
			</div>
		</div>
	);
}
