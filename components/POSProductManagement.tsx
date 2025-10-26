import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Plus, Package } from 'lucide-react'
import { api, DataTable, LoadingSpinner, ActionButtons } from '@erp-modules/shared'

interface POSProduct {
  id: string
  name: string
  sku: string
  price: number
  stock: number
  category: string
}

export function POSProductManagement() {
  const { data: products, isLoading } = useQuery({
    queryKey: ['pos-products'],
    queryFn: async () => {
      const response = await api.get<{ data: POSProduct[] }>('/api/v1/pos/products')
      return response.data.data
    },
  })

  const columns = [
    {
      accessorKey: 'name',
      header: 'Product',
      cell: ({ row }: any) => (
        <div className="flex items-center">
          <Package className="h-4 w-4 text-blue-600 mr-2" />
          <span>{row.getValue('name')}</span>
        </div>
      ),
    },
    {
      accessorKey: 'sku',
      header: 'SKU',
      cell: ({ row }: any) => <span className="font-mono text-sm">{row.getValue('sku')}</span>,
    },
    {
      accessorKey: 'price',
      header: 'Price',
      cell: ({ row }: any) => <span className="font-semibold">${row.getValue('price')}</span>,
    },
    {
      accessorKey: 'stock',
      header: 'Stock',
    },
    {
      accessorKey: 'category',
      header: 'Category',
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
        <h2 className="text-2xl font-bold text-gray-900">POS Products</h2>
        <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          <Plus className="h-4 w-4 mr-2" />
          Add Product
        </button>
      </div>
      {products && <DataTable data={products} columns={columns} />}
    </div>
  )
}

