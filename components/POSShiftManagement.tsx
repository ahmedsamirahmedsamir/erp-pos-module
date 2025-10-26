import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Clock, User } from 'lucide-react'
import { api, DataTable, LoadingSpinner, StatusBadge } from '@erp-modules/shared'

interface POSShift {
  id: string
  shift_number: string
  employee_name: string
  register_name: string
  start_time: string
  end_time: string | null
  total_sales: number
  transaction_count: number
  status: 'open' | 'closed'
}

export function POSShiftManagement() {
  const { data: shifts, isLoading } = useQuery({
    queryKey: ['pos-shifts'],
    queryFn: async () => {
      const response = await api.get<{ data: POSShift[] }>('/api/v1/pos/shifts')
      return response.data.data
    },
  })

  const columns = [
    {
      accessorKey: 'shift_number',
      header: 'Shift #',
      cell: ({ row }: any) => (
        <div className="flex items-center">
          <Clock className="h-4 w-4 text-blue-600 mr-2" />
          <span className="font-mono">{row.getValue('shift_number')}</span>
        </div>
      ),
    },
    {
      accessorKey: 'employee_name',
      header: 'Employee',
      cell: ({ row }: any) => (
        <div className="flex items-center">
          <User className="h-4 w-4 text-gray-400 mr-2" />
          <span>{row.getValue('employee_name')}</span>
        </div>
      ),
    },
    {
      accessorKey: 'register_name',
      header: 'Register',
    },
    {
      accessorKey: 'transaction_count',
      header: 'Transactions',
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
  ]

  if (isLoading) return <LoadingSpinner />

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900">Shift Management</h2>
      {shifts && <DataTable data={shifts} columns={columns} />}
    </div>
  )
}

