import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Clock, User, DollarSign } from 'lucide-react'
import { api, DataTable, LoadingSpinner, StatusBadge, ActionButtons } from '@erp-modules/shared'

interface POSSession {
  id: string
  session_number: string
  register_name: string
  cashier_name: string
  status: 'open' | 'closed'
  opening_cash: number
  closing_cash: number
  total_sales: number
  opened_at: string
  closed_at: string | null
}

export function POSSessionList() {
  const { data: sessions, isLoading } = useQuery({
    queryKey: ['pos-sessions'],
    queryFn: async () => {
      const response = await api.get<{ data: POSSession[] }>('/api/v1/pos/sessions')
      return response.data.data
    },
  })

  const columns = [
    {
      accessorKey: 'session_number',
      header: 'Session',
      cell: ({ row }: any) => <span className="font-mono">{row.getValue('session_number')}</span>,
    },
    {
      accessorKey: 'register_name',
      header: 'Register',
    },
    {
      accessorKey: 'cashier_name',
      header: 'Cashier',
      cell: ({ row }: any) => (
        <div className="flex items-center">
          <User className="h-4 w-4 text-gray-400 mr-2" />
          <span>{row.getValue('cashier_name')}</span>
        </div>
      ),
    },
    {
      accessorKey: 'total_sales',
      header: 'Sales',
      cell: ({ row }: any) => (
        <span className="font-semibold text-green-600">${row.getValue('total_sales').toLocaleString()}</span>
      ),
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }: any) => <StatusBadge status={row.getValue('status')} />,
    },
    {
      accessorKey: 'opened_at',
      header: 'Opened',
      cell: ({ row }: any) => (
        <span className="text-sm">{new Date(row.getValue('opened_at')).toLocaleString()}</span>
      ),
    },
    {
      id: 'actions',
      header: 'Actions',
      cell: ({ row }: any) => <ActionButtons onView={() => {}} onEdit={() => {}} />,
    },
  ]

  if (isLoading) return <LoadingSpinner text="Loading sessions..." />

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900">POS Sessions</h2>
      {sessions && <DataTable data={sessions} columns={columns} />}
    </div>
  )
}

