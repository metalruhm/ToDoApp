import type { Todo } from "@/types/todo";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${BASE_URL}${path}`, {
      ...init,
      headers: { "Content-Type": "application/json", ...init?.headers },
    });
  } catch {
    // ネットワーク失敗（API停止など）
    throw new Error("サーバーに接続できません");
  }

  if (!res.ok) {
    // HTTPエラー: { error: { code, message } } を読む
    const body = await res.json().catch(() => null);
    throw new Error(body?.error?.message ?? `エラーが発生しました (${res.status})`);
  }
  if (res.status === 204) return undefined as T; // 本文なし
  return res.json() as Promise<T>;
}

export type TodoStatus = "all" | "active" | "done";

export const listTodos = (status: TodoStatus = "all") =>
  request<Todo[]>(`/todos?status=${status}`);

export const createTodo = (title: string) =>
  request<Todo>("/todos", { method: "POST", body: JSON.stringify({ title }) });

export const updateTodo = (id: number, patch: { title?: string; done?: boolean }) =>
  request<Todo>(`/todos/${id}`, { method: "PATCH", body: JSON.stringify(patch) });

export const deleteTodo = (id: number) =>
  request<void>(`/todos/${id}`, { method: "DELETE" });
