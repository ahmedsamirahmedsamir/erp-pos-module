import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Plus, Percent } from 'lucide-react'
import { api, DataTable, LoadingSpinner, StatusBadge, ActionButtons } from '@erp-modules/shared'

interface POSDiscount {
  id: string
  name: string
  code: string
  type: 'percentage' | 'fixed'
  value: number
  status: 'active' | 'inactive'
  valid_from: string
  valid_until: string
}

export function POSDiscountManagement() {
  const { data: discounts, isLoading } = useQuery({
    queryKey: ['pos-discounts'],
    queryFn: async () => {
      const response = await api.get<{ data: POSDiscount[] }>('/api/v1/pos/discounts')
      return response.data.data
    },
  })

  const columns = [
    {
      accessorKey: 'name',
      header: 'Discount',
      cell: ({ row }: any) => (
        <div className="flex items-center">
          <Percent className="h-4 w-4 text-green-600 mr-2" />
          <div>
            <div className="font-medium">{row.getValue('name')}</div>
            <div className="text-sm text-gray-500">{row.original.code}</div>
          </div>
        </div>
      ),
    },
    {
      accessorKey: 'type',
      header: 'Type',
    },
    {
      accessorKey: 'value',
      header: 'Value',
      cell: ({ row }: any) => {
        const type = row.original.type
        const value = row.getValue('value')
        return type === 'percentage' ? `${value}%` : `$${value}`
      },
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }: any) => <StatusBadge status={row.getValue('status')} />,
    },
    {
      accessorKey: 'valid_until',
      header: 'Valid Until',
      cell: ({ row }: any) => (
        <span className="text-sm">{new Date(row.getValue('valid_until')).toLocaleDateString()}</span>
      ),
    },
    {
      id: 'actions',
      header: 'Actions',
      cell: ({ row }: any) => <ActionButtons onEdit={() => {}} onDelete={() => {}} />,
    },
  ]

  if (isLoading) return <LoadingSpinner />

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-gray-900">Discount Management</h2>
        <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          <Plus className="h-4 w-4 mr-2" />
          Add Discount
        </button>
      </div>
      {discounts && <DataTable data={discounts} columns={columns} />}
    </div>
  )
}

