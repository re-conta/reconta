import type { Metadata } from "next";
import { Nunito, Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const nunito = Nunito({
	variable: "--font-nunito-sans",
	subsets: ["latin"],
	preload: false,
});

const geistSans = Geist({
	variable: "--font-geist-sans",
	subsets: ["latin"],
	preload: false,
});

const geistMono = Geist_Mono({
	variable: "--font-geist-mono",
	subsets: ["latin"],
	preload: false,
});

export const metadata: Metadata = {
	title: "ReConta — Controle Financeiro Pessoal",
	description:
		"Gerencie suas finanças, analise extratos e acompanhe seus gastos com ReConta.",
	openGraph: {
		title: "ReConta — Controle Financeiro Pessoal",
		description:
			"Gerencie suas finanças, analise extratos e acompanhe seus gastos com ReConta.",
		url: "https://reconta.app",
		siteName: "ReConta",
		images: [
			{
				url: "https://reconta.app/images/coin.png",
				width: 256,
				height: 256,
				alt: "ReConta",
			},
		],
		locale: "pt-BR",
		type: "website",
	},
};

export default function RootLayout({
	children,
}: Readonly<{
	children: React.ReactNode;
}>) {
	return (
		<html
			lang="pt-BR"
			className={`${nunito.variable} ${geistSans.variable} ${geistMono.variable} dark`}
		>
			<head>
				<link
					rel="icon"
					href="/images/favicon.svg"
					sizes="any"
					type="image/svg+xml"
				/>
			</head>
			<body className="antialiased bg-zinc-950 text-zinc-100">{children}</body>
		</html>
	);
}
