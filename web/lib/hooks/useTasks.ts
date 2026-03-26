'use client'

import { useCallback, useEffect, useReducer, useRef } from 'react'
import { createClient } from '@/lib/supabase/client'
import { useRealtime } from './useRealtime'
import type { Task } from '@/lib/supabase/types'

type TaskAction =
  | { type: 'SET'; tasks: Task[] }
  | { type: 'INSERT'; task: Task }
  | { type: 'UPDATE'; task: Task }
  | { type: 'DELETE'; id: string }

function reducer(state: Task[], action: TaskAction): Task[] {
  switch (action.type) {
    case 'SET':
      return action.tasks
    case 'INSERT':
      return [...state, action.task].sort((a, b) => (a.position ?? 0) - (b.position ?? 0))
    case 'UPDATE':
      return state.map(t => t.id === action.task.id ? action.task : t)
    case 'DELETE':
      return state.filter(t => t.id !== action.id)
    default:
      return state
  }
}

export function useTasks(initialTasks: Task[], userId: string) {
  const [tasks, dispatch] = useReducer(reducer, initialTasks)
  const supabase = createClient()

  // Stable ref — callbacks read from here instead of closing over `tasks`
  const tasksRef = useRef(tasks)
  useEffect(() => { tasksRef.current = tasks }, [tasks])

  // Real-time subscription — filtered to this user's rows only
  useRealtime<Record<string, unknown>>({
    table: 'tasks',
    filter: `user_id=eq.${userId}`,
    onInsert: (row) => dispatch({ type: 'INSERT', task: row as unknown as Task }),
    onUpdate: (row) => dispatch({ type: 'UPDATE', task: row as unknown as Task }),
    onDelete: (row) => dispatch({ type: 'DELETE', id: (row as { id: string }).id }),
  })

  const addTask = useCallback(async (payload: Omit<Task, 'id' | 'user_id' | 'created_at' | 'pomodoros_done' | 'completed' | 'completed_at'>) => {
    const optimistic: Task = {
      ...payload,
      id: crypto.randomUUID(),
      user_id: userId,
      pomodoros_done: 0,
      completed: false,
      completed_at: null,
      created_at: new Date().toISOString(),
    }
    dispatch({ type: 'INSERT', task: optimistic })

    const { data, error } = await supabase.from('tasks').insert({ ...payload, user_id: userId }).select().single()
    if (error) {
      dispatch({ type: 'DELETE', id: optimistic.id })
      return { error: error.message }
    }
    // Replace optimistic row with real row from server
    dispatch({ type: 'UPDATE', task: data })
    return { data }
  }, [supabase, userId])

  const updateTask = useCallback(async (id: string, changes: Partial<Task>) => {
    const prev = tasksRef.current.find(t => t.id === id)
    if (!prev) return { error: 'Task not found' }
    dispatch({ type: 'UPDATE', task: { ...prev, ...changes } })

    const { error } = await supabase.from('tasks').update(changes).eq('id', id)
    if (error) {
      dispatch({ type: 'UPDATE', task: prev }) // rollback
      return { error: error.message }
    }
    return {}
  }, [supabase])

  const deleteTask = useCallback(async (id: string) => {
    const prev = tasksRef.current.find(t => t.id === id)
    if (!prev) return
    dispatch({ type: 'DELETE', id })

    const { error } = await supabase.from('tasks').delete().eq('id', id)
    if (error) {
      dispatch({ type: 'INSERT', task: prev }) // rollback
      return { error: error.message }
    }
    return {}
  }, [supabase])

  const completeTask = useCallback(async (id: string) => {
    const task = tasksRef.current.find(t => t.id === id)
    if (!task) return
    const now = new Date().toISOString()
    const result = await updateTask(id, { completed: true, completed_at: now })
    if (result?.error) return
    // Record completion for stats only if the update succeeded
    await supabase.from('completions').insert({
      user_id: userId,
      task_id: id,
      pomodoros_est: task.pomodoros_est,
      pomodoros_actual: task.pomodoros_done,
    })
  }, [supabase, userId, updateTask])

  return { tasks, addTask, updateTask, deleteTask, completeTask }
}
