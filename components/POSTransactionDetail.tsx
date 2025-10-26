import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Receipt } from 'lucide-react'
import { api, LoadingSpinner } from '@erp-modules/shared'

export function POSTransactionDetail({ transactionId }: { transactionId: string }) {
  const { data: transaction, isLoading } = useQuery({
    queryKey: ['pos-transaction', transactionId],
    queryFn: async () => {
      const response = await api.get(`/api/v1/pos/transactions/${transactionId}`)
      return response.data.data
    },
  })

  if (isLoading) return <LoadingSpinner />

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-2xl font-bold mb-4">Transaction Details</h2>
        <p className="font-mono">{transaction?.transaction_number}</p>
      </div>
    </div>
  )
}

