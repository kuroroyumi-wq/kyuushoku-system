const API_BASE = '/api';

const getAuthHeaders = (): Record<string, string> => {
  const token = localStorage.getItem('token');
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return headers;
};

export interface Dish {
  id: number;
  name: string;
  menu_category: string;
  serving_size: number;
  note: string;
}

export interface MenuItem {
  date: string;
  weekday?: string;
  staple?: string;
  main?: string;
  side?: string;
  soup?: string;
  dessert?: string;
  energy?: number;
  staple_id?: number;
  main_id?: number;
  side_id?: number;
  soup_id?: number;
  dessert_id?: number;
}

export interface Nutrition {
  date: string;
  energy: number;
  protein: number;
  fat: number;
  carbohydrate: number;
  salt: number;
  targets: Record<string, number>;
}

export interface OrderItem {
  ingredient_id: number;
  name: string;
  total_g: number;
}

export interface OrderResult {
  start: string;
  end: string;
  people: number;
  items: OrderItem[];
}

export interface LoginResult {
  token: string;
  user: { id: number; email: string; facility_id: number; role: string };
}

export async function login(email: string, password: string): Promise<LoginResult> {
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error || 'ログインに失敗しました');
  }
  return res.json();
}

export async function register(email: string, password: string): Promise<void> {
  const res = await fetch(`${API_BASE}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error || '登録に失敗しました');
  }
}

export function logout(): void {
  localStorage.removeItem('token');
  localStorage.removeItem('user');
}

export function isLoggedIn(): boolean {
  return !!localStorage.getItem('token');
}

export async function getDishes(category?: string): Promise<Dish[]> {
  const url = category ? `${API_BASE}/dishes?category=${encodeURIComponent(category)}` : `${API_BASE}/dishes`;
  const res = await fetch(url, { headers: getAuthHeaders() });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function getMenusByMonth(month: string): Promise<MenuItem[]> {
  const res = await fetch(`${API_BASE}/menus?month=${encodeURIComponent(month)}`, { headers: getAuthHeaders() });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function getMenuByDate(date: string): Promise<MenuItem> {
  const res = await fetch(`${API_BASE}/menus?date=${encodeURIComponent(date)}`, { headers: getAuthHeaders() });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error || res.statusText);
  }
  return res.json();
}

export async function createMenu(date: string, stapleId: number, mainId: number, sideId: number, soupId: number, dessertId: number): Promise<void> {
  const res = await fetch(`${API_BASE}/menus`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({
      date,
      staple_id: stapleId,
      main_id: mainId,
      side_id: sideId,
      soup_id: soupId,
      dessert_id: dessertId,
    }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || res.statusText);
  }
}

export async function updateMenu(date: string, stapleId: number, mainId: number, sideId: number, soupId: number, dessertId: number): Promise<void> {
  const res = await fetch(`${API_BASE}/menus`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify({
      date,
      staple_id: stapleId,
      main_id: mainId,
      side_id: sideId,
      soup_id: soupId,
      dessert_id: dessertId,
    }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || res.statusText);
  }
}

export async function getCalc(date: string): Promise<Nutrition> {
  const res = await fetch(`${API_BASE}/calc?date=${encodeURIComponent(date)}`, { headers: getAuthHeaders() });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function getOrder(start: string, end: string, people: number, excludeCondiments?: boolean): Promise<OrderResult> {
  const params = new URLSearchParams({ start, end, people: String(people) })
  if (excludeCondiments) params.set('exclude_condiments', '1')
  const res = await fetch(`${API_BASE}/order?${params}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

export interface BulkOrderItem {
  ingredient_id: number
  name: string
  total_g: number
  order_unit_g: number
  order_unit_name: string
  bulk_category: string
  order_qty: number
}

export interface BulkOrderResult {
  start: string
  end: string
  people: number
  items: BulkOrderItem[]
}

export async function getBulkOrder(start: string, end: string, people: number): Promise<BulkOrderResult> {
  const params = new URLSearchParams({ start, end, people: String(people) })
  const res = await fetch(`${API_BASE}/order/bulk?${params}`, { headers: getAuthHeaders() })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

export async function downloadExport(month: string): Promise<Blob> {
  const res = await fetch(`${API_BASE}/export?month=${encodeURIComponent(month)}`, { headers: getAuthHeaders() });
  if (!res.ok) throw new Error(await res.text());
  return res.blob();
}
