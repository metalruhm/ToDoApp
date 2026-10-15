"use client";

import { useEffect, useState } from "react";
import type { Todo } from "@/types/todo";
import { TodoForm } from "@/components/TodoForm";
import { TodoList } from "@/components/TodoList";
import { createTodo, deleteTodo, listTodos, updateTodo } from "@/lib/api";

export default function Home() {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    listTodos()
      .then(setTodos)
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false));
  }, []); // [] = 最初の1回だけ実行

  // 共通: 成功したらエラーを消し、失敗したらメッセージを出す
  async function run(action: () => Promise<void>) {
    setError("");
    try {
      await action();
    } catch (e) {
      setError(e instanceof Error ? e.message : "エラーが発生しました");
    }
  }

  const addTodo = (title: string) =>
    run(async () => {
      const created = await createTodo(title);
      setTodos((prev) => [created, ...prev]);
    });

  const toggleTodo = (id: number) =>
    run(async () => {
      const target = todos.find((t) => t.id === id);
      if (!target) return;
      const updated = await updateTodo(id, { done: !target.done });
      setTodos((prev) => prev.map((t) => (t.id === id ? updated : t)));
    });

  const editTodo = (id: number, title: string) =>
    run(async () => {
      const updated = await updateTodo(id, { title });
      setTodos((prev) => prev.map((t) => (t.id === id ? updated : t)));
    });

  const removeTodo = (id: number) =>
    run(async () => {
      await deleteTodo(id);
      setTodos((prev) => prev.filter((t) => t.id !== id));
    });

  return (
    <main>
      <h1>ToDo</h1>
      <TodoForm onAdd={addTodo} />
      {error && <p role="alert">{error}</p>}
      {loading ? (
        <p>読み込み中…</p>
      ) : (
        <TodoList todos={todos} onToggle={toggleTodo} onDelete={removeTodo} onEdit={editTodo} />
      )}
    </main>
  );
}
