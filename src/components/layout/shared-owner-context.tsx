"use client";

import type React from "react";
import { createContext, useContext, useEffect, useState } from "react";

interface SharedOwnerInfo {
	ownerId: string;
	ownerName: string;
	ownerEmail: string;
}

interface SharedOwnerContextValue {
	owner: SharedOwnerInfo;
	/** API base path: `/api/shared/{ownerId}` */
	apiBase: string;
	/** Route base path: `/compartilhado/{ownerId}` */
	routeBase: string;
}

const SharedOwnerContext = createContext<SharedOwnerContextValue | null>(null);

interface SharedOwnerProviderProps {
	ownerId: string;
	children: React.ReactNode;
}

export function SharedOwnerProvider({
	ownerId,
	children,
}: SharedOwnerProviderProps) {
	const [owner, setOwner] = useState<SharedOwnerInfo>({
		ownerId,
		ownerName: "",
		ownerEmail: "",
	});

	useEffect(() => {
		fetch(`/api/shared/${ownerId}/info`)
			.then((r) => {
				if (!r.ok) return null;
				return r.json();
			})
			.then((d) => {
				if (d) {
					setOwner({
						ownerId,
						ownerName: d.ownerName,
						ownerEmail: d.ownerEmail,
					});
				}
			});
	}, [ownerId]);

	return (
		<SharedOwnerContext.Provider
			value={{
				owner,
				apiBase: `/api/shared/${ownerId}`,
				routeBase: `/compartilhado/${ownerId}`,
			}}
		>
			{children}
		</SharedOwnerContext.Provider>
	);
}

/**
 * Returns shared owner context if inside a shared view, or null otherwise.
 */
export function useSharedOwner() {
	return useContext(SharedOwnerContext);
}
