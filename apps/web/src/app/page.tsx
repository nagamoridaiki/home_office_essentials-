"use client";

import { useEffect, useState } from "react";

type Todo = {
  id: number;
  title: string;
  completed: boolean;
};

export default function TodoPage() {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [title, setTitle] = useState("");

  const fetchTodos = async () => {
    const res = await fetch("http://localhost:8080/api/todos");
    if (res.ok) {
      const data = await res.json();
      setTodos(data || []);
    }
  };

  useEffect(() => {
    fetchTodos();
  }, []);

  const addTodo = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;

    const res = await fetch("http://localhost:8080/api/todos", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title }),
    });

    if (res.ok) {
      setTitle("");
      fetchTodos();
    }
  };

  const deleteTodo = async (id: number) => {
    await fetch(`http://localhost:8080/api/todos/${id}`, {
      method: "DELETE",
    });
    fetchTodos();
  };

  return (
    <main className="max-w-md mx-auto mt-12 p-6 bg-white rounded-xl shadow-md border border-gray-100">
      <h1 className="text-2xl font-bold text-gray-800 mb-6 text-center">Todo App</h1>
      
      <form onSubmit={addTodo} className="flex gap-2 mb-6">
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="新しいタスクを入力..."
          className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <button
          type="submit"
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg transition"
        >
          追加
        </button>
      </form>

      <ul className="space-y-2">
        {todos.map((todo) => (
          <li
            key={todo.id}
            className="flex items-center justify-between p-3 bg-gray-50 rounded-lg border border-gray-200"
          >
            <span className="text-gray-700">{todo.title}</span>
            <button
              onClick={() => deleteTodo(todo.id)}
              className="text-red-500 hover:text-red-700 text-sm font-medium"
            >
              削除
            </button>
          </li>
        ))}
      </ul>
    </main>
  );
}
