import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Plus, Monitor } from 'lucide-react'
import { api, DataTable, LoadingSpinner, StatusBadge, ActionButtons } from '@erp-modules/shared'

interface POSRegister {
  id: string
  name: string
  code: string
  location: string
  status: 'active' | 'inactive' | 'maintenance'
  current_session_id: string | null
  created_at: string
}

export function POSRegisterManagement() {
  const { data: registers, isLoading } = useQuery({
    queryKey: ['pos-registers'],
    queryFn: async () => {
      const response = await api.get<{ data: POSRegister[] }>('/api/v1/pos/registers')
      return response.data.data
    },
  })

  const columns = [
    {
      accessorKey: 'name',
      header: 'Register',
      cell: ({ row }: any) => (
        <div className="flex items-center">
          <Monitor className="h-4 w-4 text-blue-600 mr-2" />
          <div>
            <div className="font-medium">{row.getValue('name')}</div>
            <div className="text-sm text-gray-500">{row.original.code}</div>
          </div>
        </div>
      ),
    },
    {
      accessorKey: 'location',
      header: 'Location',
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
        <h2 className="text-2xl font-bold text-gray-900">POS Registers</h2>
        <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          <Plus className="h-4 w-4 mr-2" />
          Add Register
        </button>
      </div>
      {registers && <DataTable data={registers} columns={columns} />}
    </div>
  )
}

