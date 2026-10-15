"use client";

import { useState } from "react";
import type { Todo } from "@/types/todo";
import { TodoForm } from "@/components/TodoForm";
import { TodoList } from "@/components/TodoList";

export default function Home() {
  const [todos, setTodos] = useState<Todo[]>([]);

  function addTodo(title: string) {
    const now = new Date().toISOString();
    const todo: Todo = { id: Date.now(), title, done: false, createdAt: now, updatedAt: now };
    setTodos((prev) => [todo, ...prev]); // 元の配列を書き換えず、新しい配列を作る
  }

  function toggleTodo(id: number) {
    setTodos((prev) => prev.map((t) => (t.id === id ? { ...t, done: !t.done } : t)));
  }

  function deleteTodo(id: number) {
    setTodos((prev) => prev.filter((t) => t.id !== id));
  }

  function editTodo(id: number, title: string) {
    setTodos((prev) => prev.map((t) => (t.id === id ? { ...t, title } : t)));
  }

  return (
    <main>
      <h1>ToDo</h1>
      <TodoForm onAdd={addTodo} />
      <TodoList todos={todos} onToggle={toggleTodo} onDelete={deleteTodo} onEdit={editTodo} />
    </main>
  );
}
