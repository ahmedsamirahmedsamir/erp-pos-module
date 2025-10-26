import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { FileText, Download, Printer } from 'lucide-react'
import { api, DataTable, LoadingSpinner } from '@erp-modules/shared'

interface POSReceipt {
  id: string
  receipt_number: string
  transaction_id: string
  customer_name: string
  total_amount: number
  created_at: string
}

export function POSReceiptList() {
  const { data: receipts, isLoading } = useQuery({
    queryKey: ['pos-receipts'],
    queryFn: async () => {
      const response = await api.get<{ data: POSReceipt[] }>('/api/v1/pos/receipts')
      return response.data.data
    },
  })

  const columns = [
    {
      accessorKey: 'receipt_number',
      header: 'Receipt #',
      cell: ({ row }: any) => (
        <div className="flex items-center">
          <FileText className="h-4 w-4 text-blue-600 mr-2" />
          <span className="font-mono">{row.getValue('receipt_number')}</span>
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
      accessorKey: 'created_at',
      header: 'Date',
      cell: ({ row }: any) => (
        <span className="text-sm">{new Date(row.getValue('created_at')).toLocaleString()}</span>
      ),
    },
    {
      id: 'actions',
      header: 'Actions',
      cell: ({ row }: any) => (
        <div className="flex items-center space-x-2">
          <button className="p-1 text-blue-600 hover:text-blue-800" title="Download">
            <Download className="h-4 w-4" />
          </button>
          <button className="p-1 text-green-600 hover:text-green-800" title="Print">
            <Printer className="h-4 w-4" />
          </button>
        </div>
      ),
    },
  ]

  if (isLoading) return <LoadingSpinner text="Loading receipts..." />

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900">POS Receipts</h2>
      {receipts && <DataTable data={receipts} columns={columns} />}
    </div>
  )
}

