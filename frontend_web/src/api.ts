const API_BASE = '/api';

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

export async function getDishes(category?: string): Promise<Dish[]> {
  const url = category ? `${API_BASE}/dishes?category=${encodeURIComponent(category)}` : `${API_BASE}/dishes`;
  const res = await fetch(url);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function getMenusByMonth(month: string): Promise<MenuItem[]> {
  const res = await fetch(`${API_BASE}/menus?month=${encodeURIComponent(month)}`);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function getMenuByDate(date: string): Promise<MenuItem> {
  const res = await fetch(`${API_BASE}/menus?date=${encodeURIComponent(date)}`);
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error || res.statusText);
  }
  return res.json();
}

export async function createMenu(date: string, stapleId: number, mainId: number, sideId: number, soupId: number, dessertId: number): Promise<void> {
  const res = await fetch(`${API_BASE}/menus`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
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
    headers: { 'Content-Type': 'application/json' },
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
  const res = await fetch(`${API_BASE}/calc?date=${encodeURIComponent(date)}`);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function getOrder(start: string, end: string, people: number): Promise<OrderResult> {
  const res = await fetch(`${API_BASE}/order?start=${start}&end=${end}&people=${people}`);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function downloadExport(month: string): Promise<Blob> {
  const res = await fetch(`${API_BASE}/export?month=${encodeURIComponent(month)}`);
  if (!res.ok) throw new Error(await res.text());
  return res.blob();
}
