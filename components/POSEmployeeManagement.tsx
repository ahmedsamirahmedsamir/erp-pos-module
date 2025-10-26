import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Plus, User, Badge } from 'lucide-react'
import { api, DataTable, LoadingSpinner, StatusBadge, ActionButtons } from '@erp-modules/shared'

interface POSEmployee {
  id: string
  name: string
  email: string
  phone: string
  role: string
  pin_code: string
  status: 'active' | 'inactive'
  hire_date: string
}

export function POSEmployeeManagement() {
  const { data: employees, isLoading } = useQuery({
    queryKey: ['pos-employees'],
    queryFn: async () => {
      const response = await api.get<{ data: POSEmployee[] }>('/api/v1/pos/employees')
      return response.data.data
    },
  })

  const columns = [
    {
      accessorKey: 'name',
      header: 'Employee',
      cell: ({ row }: any) => (
        <div className="flex items-center">
          <User className="h-4 w-4 text-blue-600 mr-2" />
          <span>{row.getValue('name')}</span>
        </div>
      ),
    },
    {
      accessorKey: 'email',
      header: 'Email',
    },
    {
      accessorKey: 'phone',
      header: 'Phone',
    },
    {
      accessorKey: 'role',
      header: 'Role',
      cell: ({ row }: any) => (
        <div className="flex items-center">
          <Badge className="h-4 w-4 text-purple-600 mr-2" />
          <span>{row.getValue('role')}</span>
        </div>
      ),
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
        <h2 className="text-2xl font-bold text-gray-900">Employee Management</h2>
        <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          <Plus className="h-4 w-4 mr-2" />
          Add Employee
        </button>
      </div>
      {employees && <DataTable data={employees} columns={columns} />}
    </div>
  )
}

