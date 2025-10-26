import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { CreditCard } from 'lucide-react'
import { api, DataTable, LoadingSpinner, StatusBadge } from '@erp-modules/shared'

interface POSPayment {
  id: string
  payment_number: string
  transaction_id: string
  amount: number
  payment_method: string
  status: 'completed' | 'pending' | 'failed'
  created_at: string
}

export function POSPaymentList() {
  const { data: payments, isLoading } = useQuery({
    queryKey: ['pos-payments'],
    queryFn: async () => {
      const response = await api.get<{ data: POSPayment[] }>('/api/v1/pos/payments')
      return response.data.data
    },
  })

  const columns = [
    {
      accessorKey: 'payment_number',
      header: 'Payment #',
      cell: ({ row }: any) => (
        <div className="flex items-center">
          <CreditCard className="h-4 w-4 text-green-600 mr-2" />
          <span className="font-mono">{row.getValue('payment_number')}</span>
        </div>
      ),
    },
    {
      accessorKey: 'amount',
      header: 'Amount',
      cell: ({ row }: any) => (
        <span className="font-semibold">${row.getValue('amount').toLocaleString()}</span>
      ),
    },
    {
      accessorKey: 'payment_method',
      header: 'Method',
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }: any) => <StatusBadge status={row.getValue('status')} />,
    },
    {
      accessorKey: 'created_at',
      header: 'Date',
      cell: ({ row }: any) => (
        <span className="text-sm">{new Date(row.getValue('created_at')).toLocaleString()}</span>
      ),
    },
  ]

  if (isLoading) return <LoadingSpinner text="Loading payments..." />

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900">POS Payments</h2>
      {payments && <DataTable data={payments} columns={columns} />}
    </div>
  )
}

