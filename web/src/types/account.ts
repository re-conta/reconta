export interface Account {
  id: number;
  name: string;
  type: string;
  balance: number;
  createdAt: string;
}

export interface AccountInput {
  name: string;
  type: string;
  balance: number;
}
