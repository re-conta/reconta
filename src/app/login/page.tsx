"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";
import { signIn } from "@/lib/auth-client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

function LoginForm() {
	const router = useRouter();
	const searchParams = useSearchParams();
	const next = searchParams.get("next") ?? "/";

	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [error, setError] = useState<string | null>(null);
	const [loading, setLoading] = useState(false);

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		setLoading(true);
		setError(null);

		const { error } = await signIn.email({ email, password });

		if (error) {
			setError("E-mail ou senha incorretos.");
			setLoading(false);
			return;
		}

		router.push(next);
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
					<h1 className="text-xl font-semibold text-zinc-100 mb-1">Entrar</h1>
					<p className="text-sm text-zinc-400 mb-6">
						Acesse sua conta para continuar
					</p>

					<form onSubmit={handleSubmit} className="space-y-4">
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
								placeholder="••••••••"
								value={password}
								onChange={(e) => setPassword(e.target.value)}
								required
								autoComplete="current-password"
								className="mt-1"
							/>
						</div>

						{error && (
							<p className="text-sm text-red-400 bg-red-900/20 border border-red-800/30 rounded-lg px-3 py-2">
								{error}
							</p>
						)}

						<Button type="submit" className="w-full" disabled={loading}>
							{loading ? "Entrando..." : "Entrar"}
						</Button>
					</form>

					<p className="text-center text-sm text-zinc-500 mt-6">
						Não tem uma conta?{" "}
						<Link href="/cadastro" className="text-indigo-400 hover:underline">
							Criar conta
						</Link>
					</p>
				</div>
			</div>
		</div>
	);
}

export default function LoginPage() {
	return (
		<Suspense>
			<LoginForm />
		</Suspense>
	);
}
