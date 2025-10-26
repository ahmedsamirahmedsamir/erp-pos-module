import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Plus, User } from 'lucide-react'
import { api, DataTable, LoadingSpinner, ActionButtons } from '@erp-modules/shared'

interface POSCustomer {
  id: string
  name: string
  email: string
  phone: string
  loyalty_points: number
  total_purchases: number
}

export function POSCustomerManagement() {
  const { data: customers, isLoading } = useQuery({
    queryKey: ['pos-customers'],
    queryFn: async () => {
      const response = await api.get<{ data: POSCustomer[] }>('/api/v1/pos/customers')
      return response.data.data
    },
  })

  const columns = [
    {
      accessorKey: 'name',
      header: 'Customer',
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
      accessorKey: 'loyalty_points',
      header: 'Points',
      cell: ({ row }: any) => (
        <span className="font-medium text-purple-600">{row.getValue('loyalty_points')}</span>
      ),
    },
    {
      accessorKey: 'total_purchases',
      header: 'Purchases',
      cell: ({ row }: any) => (
        <span className="font-semibold">${row.getValue('total_purchases').toLocaleString()}</span>
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
        <h2 className="text-2xl font-bold text-gray-900">POS Customers</h2>
        <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          <Plus className="h-4 w-4 mr-2" />
          Add Customer
        </button>
      </div>
      {customers && <DataTable data={customers} columns={columns} />}
    </div>
  )
}

