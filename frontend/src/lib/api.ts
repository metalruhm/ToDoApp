import type { Todo } from "@/types/todo";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${BASE_URL}${path}`, {
      ...init,
      credentials: "include", // Cookie を送受信する
      headers: { "Content-Type": "application/json", ...init?.headers },
    });
  } catch {
    // ネットワーク失敗（API停止など）
    throw new ApiError("サーバーに接続できません", 0);
  }

  if (!res.ok) {
    // HTTPエラー: { error: { code, message } } を読む
    const body = await res.json().catch(() => null);
    throw new ApiError(
      body?.error?.message ?? `エラーが発生しました (${res.status})`,
      res.status,
    );
  }
  if (res.status === 204) return undefined as T; // 本文なし
  return res.json() as Promise<T>;
}

export type TodoStatus = "all" | "active" | "done";

export const listTodos = (status: TodoStatus = "all") =>
  request<Todo[]>(`/todos?status=${status}`);

export const createTodo = (title: string) =>
  request<Todo>("/todos", { method: "POST", body: JSON.stringify({ title }) });

export const updateTodo = (
  id: number,
  patch: { title?: string; done?: boolean },
) =>
  request<Todo>(`/todos/${id}`, {
    method: "PATCH",
    body: JSON.stringify(patch),
  });

export const deleteTodo = (id: number) =>
  request<void>(`/todos/${id}`, { method: "DELETE" });

export type User = { id: number; email: string };

export const register = (email: string, password: string) =>
  request<User>("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });

export const login = (email: string, password: string) =>
  request<User>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });

export const logout = () => request<void>("/auth/logout", { method: "POST" });

export const me = () => request<User>("/auth/me");
