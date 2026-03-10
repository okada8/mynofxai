import { useState, useEffect, useRef, useCallback } from 'react'

// Shallow comparison helper
function shallowEqual(objA: any, objB: any): boolean {
  if (Object.is(objA, objB)) return true

  if (
    typeof objA !== 'object' ||
    objA === null ||
    typeof objB !== 'object' ||
    objB === null
  ) {
    return false
  }

  const keysA = Object.keys(objA)
  const keysB = Object.keys(objB)

  if (keysA.length !== keysB.length) return false

  for (let i = 0; i < keysA.length; i++) {
    if (
      !Object.prototype.hasOwnProperty.call(objB, keysA[i]) ||
      !Object.is(objA[keysA[i]], objB[keysA[i]])
    ) {
      return false
    }
  }

  return true
}

type PollingCallback<T> = () => Promise<T>

interface UseOptimizedPollingOptions {
  interval?: number
  enabled?: boolean
  onSuccess?: (data: any) => void
  onError?: (error: any) => void
}

export function useOptimizedPolling<T>(
  callback: PollingCallback<T>,
  options: UseOptimizedPollingOptions = {}
) {
  const { interval = 1000, enabled = true, onSuccess, onError } = options

  const [data, setData] = useState<T | null>(null)
  const prevDataRef = useRef<T | null>(null)
  const timeoutRef = useRef<NodeJS.Timeout | null>(null)
  const mountedRef = useRef(true)

  // Dynamic interval management
  const [currentInterval, setCurrentInterval] = useState(interval)

  useEffect(() => {
    mountedRef.current = true
    return () => {
      mountedRef.current = false
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  const poll = useCallback(async () => {
    if (!enabled || !mountedRef.current) return

    try {
      const newData = await callback()

      if (!mountedRef.current) return

      // Shallow comparison to avoid unnecessary re-renders
      if (!shallowEqual(prevDataRef.current, newData)) {
        prevDataRef.current = newData
        setData(newData)
        if (onSuccess) onSuccess(newData)
      }

      // Schedule next poll
      timeoutRef.current = setTimeout(poll, currentInterval)
    } catch (err) {
      if (!mountedRef.current) return

      if (onError) onError(err)

      // Retry with backoff or standard interval
      timeoutRef.current = setTimeout(poll, currentInterval)
    }
  }, [callback, enabled, currentInterval, onSuccess, onError])

  useEffect(() => {
    if (enabled) {
      poll()
    } else {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [enabled, poll])

  // Function to manually update interval (e.g. slow down when completed)
  const setPollingInterval = (newInterval: number) => {
    setCurrentInterval(newInterval)
  }

  return { data, setPollingInterval }
}
