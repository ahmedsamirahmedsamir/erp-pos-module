import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { DollarSign, Lock, Unlock } from 'lucide-react'
import { api, DataTable, LoadingSpinner, StatusBadge } from '@erp-modules/shared'

interface CashDrawer {
  id: string
  register_name: string
  current_cash: number
  status: 'open' | 'closed'
  last_count_at: string
}

export function POSCashDrawerManagement() {
  const { data: drawers, isLoading } = useQuery({
    queryKey: ['cash-drawers'],
    queryFn: async () => {
      const response = await api.get<{ data: CashDrawer[] }>('/api/v1/pos/cash-drawers')
      return response.data.data
    },
  })

  const columns = [
    {
      accessorKey: 'register_name',
      header: 'Register',
    },
    {
      accessorKey: 'current_cash',
      header: 'Cash',
      cell: ({ row }: any) => (
        <span className="font-semibold">${row.getValue('current_cash').toLocaleString()}</span>
      ),
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }: any) => {
        const status = row.getValue('status')
        return (
          <div className="flex items-center">
            {status === 'open' ? (
              <Unlock className="h-4 w-4 text-green-600 mr-2" />
            ) : (
              <Lock className="h-4 w-4 text-gray-600 mr-2" />
            )}
            <StatusBadge status={status} />
          </div>
        )
      },
    },
    {
      accessorKey: 'last_count_at',
      header: 'Last Count',
      cell: ({ row }: any) => (
        <span className="text-sm">{new Date(row.getValue('last_count_at')).toLocaleString()}</span>
      ),
    },
  ]

  if (isLoading) return <LoadingSpinner />

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900">Cash Drawer Management</h2>
      {drawers && <DataTable data={drawers} columns={columns} />}
    </div>
  )
}

