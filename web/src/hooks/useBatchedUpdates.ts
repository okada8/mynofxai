import { useRef, useEffect, useCallback } from 'react'

export function useBatchedUpdates<T>(
  onProcessBatch: (batch: T[]) => void,
  interval = 100
) {
  const batchRef = useRef<T[]>([])
  const timeoutRef = useRef<NodeJS.Timeout | null>(null)

  const processBatch = useCallback(() => {
    if (batchRef.current.length > 0) {
      onProcessBatch([...batchRef.current])
      batchRef.current = []
    }
    timeoutRef.current = null
  }, [onProcessBatch])

  const addUpdate = useCallback(
    (update: T) => {
      batchRef.current.push(update)

      if (!timeoutRef.current) {
        timeoutRef.current = setTimeout(processBatch, interval)
      }
    },
    [processBatch, interval]
  )

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
        // Process remaining items on unmount?
        // Usually better not to trigger state updates on unmount
      }
    }
  }, [])

  return { addUpdate }
}
