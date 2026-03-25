"use client";

import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { signIn, signUp } from "@/lib/auth-client";
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
	const [googleLoading, setGoogleLoading] = useState(false);

	async function handleGoogleSignIn() {
		setGoogleLoading(true);
		await signIn.social({ provider: "google", callbackURL: "/" });
		setGoogleLoading(false);
	}

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
					<Image
						src="/images/favicon.svg"
						alt="ReConta"
						width={48}
						height={48}
						unoptimized
					/>
					<span className="text-3xl font-bold text-white">ReConta</span>
				</div>

				<div className="bg-zinc-900 border border-zinc-800 rounded-2xl p-8">
					<h1 className="text-xl font-semibold text-zinc-100 mb-1">
						Criar conta
					</h1>
					<p className="text-sm text-zinc-400 mb-6">
						Suas finanças, organizadas e privadas
					</p>

					<Button
						type="button"
						variant="outline"
						className="w-full gap-2 mb-4"
						onClick={handleGoogleSignIn}
						disabled={googleLoading}
					>
						<svg viewBox="0 0 24 24" className="h-4 w-4" aria-hidden="true">
							<path
								d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
								fill="#4285F4"
							/>
							<path
								d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
								fill="#34A853"
							/>
							<path
								d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z"
								fill="#FBBC05"
							/>
							<path
								d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
								fill="#EA4335"
							/>
						</svg>
						{googleLoading ? "Redirecionando..." : "Continuar com Google"}
					</Button>

					<div className="flex items-center gap-3 mb-4">
						<div className="flex-1 h-px bg-zinc-800" />
						<span className="text-xs text-zinc-500">ou</span>
						<div className="flex-1 h-px bg-zinc-800" />
					</div>

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
