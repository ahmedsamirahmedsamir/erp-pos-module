import React from 'react'
import { ShoppingCart, DollarSign, Users, TrendingUp, Receipt, CreditCard, Package, BarChart3 } from 'lucide-react'
import { ModuleDashboard, useModuleQuery } from '@erp-modules/shared'
import { OverviewTab } from './tabs/OverviewTab'
import { TerminalTab } from './tabs/TerminalTab'
import { TransactionsTab } from './tabs/TransactionsTab'
import { SessionsTab } from './tabs/SessionsTab'
import { RegistersTab } from './tabs/RegistersTab'
import { ProductsTab } from './tabs/ProductsTab'
import { CustomersTab } from './tabs/CustomersTab'
import { ReportsTab } from './tabs/ReportsTab'

interface POSAnalytics {
  today_sales: number
  today_transactions: number
  today_customers: number
  average_transaction_value: number
}

export default function POSDashboard() {
  const { data: analytics } = useModuleQuery<{ data: POSAnalytics }>(
    ['pos-analytics'],
    '/api/v1/pos/analytics'
  )

  const analyticsData = analytics?.data

  return (
    <ModuleDashboard
      title="Point of Sale"
      icon={ShoppingCart}
      description="Complete POS system with transactions, sessions, and customer management"
      kpis={[
        {
          id: 'sales',
          label: "Today's Sales",
          value: `$${analyticsData?.today_sales?.toFixed(2) || 0}`,
          icon: DollarSign,
          color: 'green',
        },
        {
          id: 'transactions',
          label: "Today's Transactions",
          value: analyticsData?.today_transactions || 0,
          icon: Receipt,
          color: 'blue',
        },
        {
          id: 'customers',
          label: "Today's Customers",
          value: analyticsData?.today_customers || 0,
          icon: Users,
          color: 'purple',
        },
        {
          id: 'avg-transaction',
          label: 'Avg Transaction',
          value: `$${analyticsData?.average_transaction_value?.toFixed(2) || 0}`,
          icon: TrendingUp,
          color: 'orange',
        },
      ]}
      actions={[
        {
          id: 'open-terminal',
          label: 'Open Terminal',
          icon: ShoppingCart,
          onClick: () => window.location.href = '/pos/terminal',
          variant: 'primary',
        },
      ]}
      tabs={[
        {
          id: 'overview',
          label: 'Overview',
          icon: BarChart3,
          content: <OverviewTab analytics={analyticsData} />,
        },
        {
          id: 'terminal',
          label: 'POS Terminal',
          icon: ShoppingCart,
          content: <TerminalTab />,
        },
        {
          id: 'transactions',
          label: 'Transactions',
          icon: Receipt,
          content: <TransactionsTab />,
        },
        {
          id: 'sessions',
          label: 'Sessions',
          icon: CreditCard,
          content: <SessionsTab />,
        },
        {
          id: 'registers',
          label: 'Registers',
          icon: Package,
          content: <RegistersTab />,
        },
        {
          id: 'products',
          label: 'Products',
          icon: Package,
          content: <ProductsTab />,
        },
        {
          id: 'customers',
          label: 'Customers',
          icon: Users,
          content: <CustomersTab />,
        },
        {
          id: 'reports',
          label: 'Reports',
          icon: BarChart3,
          content: <ReportsTab />,
        },
      ]}
      defaultTab="overview"
    />
  )
}
