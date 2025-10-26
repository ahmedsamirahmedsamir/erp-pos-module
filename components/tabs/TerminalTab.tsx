import React from 'react'
import { ShoppingCart } from 'lucide-react'

export function TerminalTab() {
  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="flex items-center mb-4">
        <ShoppingCart className="h-6 w-6 mr-2 text-blue-600" />
        <h3 className="text-lg font-semibold text-gray-900">POS Terminal</h3>
      </div>
      <p className="text-gray-600 mb-4">For full terminal experience, use the dedicated POSTerminal component.</p>
      <button 
        onClick={() => window.location.href = '/pos/terminal'}
        className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
      >
        Open Full Terminal
      </button>
    </div>
  )
}

