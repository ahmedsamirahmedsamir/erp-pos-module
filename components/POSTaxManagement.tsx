import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Plus, Calculator } from 'lucide-react'
import { api, DataTable, LoadingSpinner, StatusBadge, ActionButtons } from '@erp-modules/shared'

interface POSTax {
  id: string
  name: string
  code: string
  rate: number
  type: 'sales_tax' | 'vat' | 'gst'
  status: 'active' | 'inactive'
  applies_to: string
}

export function POSTaxManagement() {
  const { data: taxes, isLoading } = useQuery({
    queryKey: ['pos-taxes'],
    queryFn: async () => {
      const response = await api.get<{ data: POSTax[] }>('/api/v1/pos/taxes')
      return response.data.data
    },
  })

  const columns = [
    {
      accessorKey: 'name',
      header: 'Tax Name',
      cell: ({ row }: any) => (
        <div className="flex items-center">
          <Calculator className="h-4 w-4 text-blue-600 mr-2" />
          <div>
            <div className="font-medium">{row.getValue('name')}</div>
            <div className="text-sm text-gray-500">{row.original.code}</div>
          </div>
        </div>
      ),
    },
    {
      accessorKey: 'rate',
      header: 'Rate',
      cell: ({ row }: any) => <span className="font-medium">{row.getValue('rate')}%</span>,
    },
    {
      accessorKey: 'type',
      header: 'Type',
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }: any) => <StatusBadge status={row.getValue('status')} />,
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
        <h2 className="text-2xl font-bold text-gray-900">Tax Management</h2>
        <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          <Plus className="h-4 w-4 mr-2" />
          Add Tax
        </button>
      </div>
      {taxes && <DataTable data={taxes} columns={columns} />}
    </div>
  )
}

