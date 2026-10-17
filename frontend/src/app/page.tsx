"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import type { Todo } from "@/types/todo";
import { TodoForm } from "@/components/TodoForm";
import { TodoList } from "@/components/TodoList";
import { StatusFilter } from "@/components/StatusFilter";
import {
  ApiError,
  createTodo,
  deleteTodo,
  listTodos,
  logout,
  me,
  updateTodo,
  type TodoStatus,
  type User,
} from "@/lib/api";

export default function Home() {
  const router = useRouter();
  const [status, setStatus] = useState<TodoStatus>("all");
  // どの絞り込み条件で取得した結果かを一緒に持つ（条件が変わったら「読み込み中」になる）
  const [data, setData] = useState<{ status: TodoStatus; todos: Todo[] } | null>(null);
  const [loadFailed, setLoadFailed] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);
  const [user, setUser] = useState<User | null>(null);
  const [error, setError] = useState("");

  const loading = !loadFailed && (data === null || data.status !== status);

  const handleError = useCallback(
    (e: unknown) => {
      if (e instanceof ApiError && e.status === 401) {
        router.replace("/login"); // 未ログインならログイン画面へ
        return;
      }
      setError(e instanceof Error ? e.message : "エラーが発生しました");
    },
    [router],
  );

  useEffect(() => {
    me().then(setUser).catch(handleError);
  }, [handleError]);

  useEffect(() => {
    let cancelled = false; // 条件が変わった後に古い応答が届いても無視する
    listTodos(status)
      .then((todos) => {
        if (cancelled) return;
        setData({ status, todos });
        setLoadFailed(false);
        setError("");
      })
      .catch((e) => {
        if (cancelled) return;
        setLoadFailed(true);
        handleError(e);
      });
    return () => {
      cancelled = true;
    };
  }, [status, reloadKey, handleError]);

  // 変更系の共通処理: 成功したら一覧を取り直す。失敗したらメッセージを出す。
  async function run(action: () => Promise<unknown>): Promise<boolean> {
    setError("");
    try {
      await action();
      setReloadKey((k) => k + 1);
      return true;
    } catch (e) {
      handleError(e);
      return false;
    }
  }

  const addTodo = (title: string) => run(() => createTodo(title));
  const toggleTodo = (id: number) => {
    const target = data?.todos.find((t) => t.id === id);
    if (target) run(() => updateTodo(id, { done: !target.done }));
  };
  const editTodo = (id: number, title: string) => run(() => updateTodo(id, { title }));
  const removeTodo = (id: number) => run(() => deleteTodo(id));

  async function handleLogout() {
    try {
      await logout();
    } finally {
      router.replace("/login");
    }
  }

  return (
    <main className="mx-auto w-full max-w-xl px-4 py-8">
      <header className="mb-6 flex items-center justify-between gap-3">
        <h1 className="text-2xl font-bold">ToDo</h1>
        <div className="flex items-center gap-3 text-sm">
          {user && <span className="truncate text-slate-500">{user.email}</span>}
          <button
            type="button"
            onClick={handleLogout}
            className="rounded px-2 py-1 text-indigo-600 underline hover:bg-indigo-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            ログアウト
          </button>
        </div>
      </header>

      <TodoForm onAdd={addTodo} />
      <StatusFilter value={status} onChange={setStatus} />

      {error && (
        <p role="alert" className="mt-3 text-sm text-red-600">
          {error}
        </p>
      )}

      {loading ? (
        <p className="mt-6 text-center text-slate-500">読み込み中…</p>
      ) : (
        data && (
          <TodoList
            todos={data.todos}
            onToggle={toggleTodo}
            onDelete={removeTodo}
            onEdit={editTodo}
          />
        )
      )}
    </main>
  );
}
