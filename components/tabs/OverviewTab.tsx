import React from 'react'
import { TrendingUp, DollarSign } from 'lucide-react'

interface OverviewTabProps {
  analytics?: {
    today_sales: number
    today_transactions: number
    today_customers: number
    average_transaction_value: number
  }
}

export function OverviewTab({ analytics }: OverviewTabProps) {
  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
          <DollarSign className="h-5 w-5 mr-2" />
          Sales Overview
        </h3>
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-green-50 p-4 rounded-lg">
            <div className="text-sm text-gray-600">Total Sales</div>
            <div className="text-2xl font-bold text-green-600">${analytics?.today_sales?.toFixed(2) || 0}</div>
          </div>
          <div className="bg-blue-50 p-4 rounded-lg">
            <div className="text-sm text-gray-600">Avg Transaction</div>
            <div className="text-2xl font-bold text-blue-600">${analytics?.average_transaction_value?.toFixed(2) || 0}</div>
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
          <TrendingUp className="h-5 w-5 mr-2" />
          Today's Activity
        </h3>
        <div className="space-y-3">
          <div className="flex justify-between items-center p-3 bg-gray-50 rounded">
            <span className="text-sm text-gray-600">Transactions</span>
            <span className="text-lg font-semibold">{analytics?.today_transactions || 0}</span>
          </div>
          <div className="flex justify-between items-center p-3 bg-gray-50 rounded">
            <span className="text-sm text-gray-600">Customers Served</span>
            <span className="text-lg font-semibold">{analytics?.today_customers || 0}</span>
          </div>
        </div>
      </div>
    </div>
  )
}

