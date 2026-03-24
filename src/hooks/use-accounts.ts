"use client";

import { useEffect, useState } from "react";

interface Account {
	id: number;
	name: string;
	type: string;
	balance: number;
}

export function useAccounts() {
	const [accounts, setAccounts] = useState<Account[]>([]);

	useEffect(() => {
		fetch("/api/accounts")
			.then((r) => r.json())
			.then(setAccounts);
	}, []);

	return { accounts };
}
