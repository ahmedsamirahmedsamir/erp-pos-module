import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Clock, DollarSign, Receipt } from 'lucide-react'
import { api, LoadingSpinner, StatusBadge } from '@erp-modules/shared'

export function POSSessionDetail({ sessionId }: { sessionId: string }) {
  const { data: session, isLoading } = useQuery({
    queryKey: ['pos-session', sessionId],
    queryFn: async () => {
      const response = await api.get(`/api/v1/pos/sessions/${sessionId}`)
      return response.data.data
    },
  })

  if (isLoading) return <LoadingSpinner />

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-2xl font-bold mb-4">Session Details</h2>
        <dl className="grid grid-cols-2 gap-4">
          <div>
            <dt className="text-sm text-gray-500">Session Number</dt>
            <dd className="font-mono font-medium">{session?.session_number}</dd>
          </div>
          <div>
            <dt className="text-sm text-gray-500">Status</dt>
            <dd><StatusBadge status={session?.status} /></dd>
          </div>
        </dl>
      </div>
    </div>
  )
}

