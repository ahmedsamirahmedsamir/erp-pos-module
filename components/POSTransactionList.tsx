import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Receipt } from 'lucide-react'
import { api, DataTable, LoadingSpinner, StatusBadge, ActionButtons } from '@erp-modules/shared'

interface POSTransaction {
  id: string
  transaction_number: string
  customer_name: string
  total_amount: number
  payment_method: string
  status: 'completed' | 'cancelled' | 'refunded'
  transaction_date: string
}

export function POSTransactionList() {
  const { data: transactions, isLoading } = useQuery({
    queryKey: ['pos-transactions'],
    queryFn: async () => {
      const response = await api.get<{ data: POSTransaction[] }>('/api/v1/pos/transactions')
      return response.data.data
    },
  })

  const columns = [
    {
      accessorKey: 'transaction_number',
      header: 'Transaction #',
      cell: ({ row }: any) => (
        <div className="flex items-center">
          <Receipt className="h-4 w-4 text-blue-600 mr-2" />
          <span className="font-mono">{row.getValue('transaction_number')}</span>
        </div>
      ),
    },
    {
      accessorKey: 'customer_name',
      header: 'Customer',
    },
    {
      accessorKey: 'total_amount',
      header: 'Amount',
      cell: ({ row }: any) => (
        <span className="font-semibold">${row.getValue('total_amount').toLocaleString()}</span>
      ),
    },
    {
      accessorKey: 'payment_method',
      header: 'Payment',
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }: any) => <StatusBadge status={row.getValue('status')} />,
    },
    {
      accessorKey: 'transaction_date',
      header: 'Date',
      cell: ({ row }: any) => (
        <span className="text-sm">{new Date(row.getValue('transaction_date')).toLocaleString()}</span>
      ),
    },
    {
      id: 'actions',
      header: 'Actions',
      cell: ({ row }: any) => <ActionButtons onView={() => {}} />,
    },
  ]

  if (isLoading) return <LoadingSpinner text="Loading transactions..." />

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900">POS Transactions</h2>
      {transactions && <DataTable data={transactions} columns={columns} pageSize={20} />}
    </div>
  )
}

