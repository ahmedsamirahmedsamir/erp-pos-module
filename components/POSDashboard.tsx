import React, { useState, useEffect, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { 
  Plus, Minus, Trash2, CreditCard, DollarSign, Receipt, User, Search,
  ShoppingCart, Package, Calculator, Percent, Hash, Tag, Barcode, ScanLine,
  Camera, QrCode, Smartphone, Monitor, Tablet, Laptop, Headphones,
  Mic, Speaker, Printer, Scanner, Fax, Wifi, WifiOff, Signal, Battery,
  Thermometer, Droplets, Sun, Moon, Wind, Snowflake, MapPin, Truck,
  Users, Phone, Mail, Globe, Settings, Zap, Bell, Activity, Target,
  TrendingUp, TrendingDown, Calendar, AlertTriangle, CheckCircle, Clock,
  Download, Upload, RefreshCw, Play, Pause, MoreHorizontal, Star, X,
  ArrowUpRight, ArrowDownRight, ArrowRight, ArrowLeft, Eye, Edit,
  FileText, PieChart, LineChart, BarChart3, Database, Cloud,
  CheckSquare, Square, Clipboard, Handshake, Briefcase, Building, Home,
  Award, Gift, Layers, RotateCcw, Archive, AlertCircle, CheckSquare as CheckSquareIcon,
  ArrowUpDown, ArrowDownUp, ArrowRightLeft, ArrowLeftRight, FileSpreadsheet,
  FileImage, FileVideo, FileAudio, Smartphone as SmartphoneIcon, Monitor as MonitorIcon,
  Tablet as TabletIcon, Laptop as LaptopIcon, Headphones as HeadphonesIcon
} from 'lucide-react'
import { api } from '../../lib/api'

interface Product {
  id: string
  name: string
  sku: string
  price: number
  cost: number
  category: string
  brand: string
  barcode: string
  qr_code: string
  image_url: string
  description: string
  stock_quantity: number
  min_stock_level: number
  tax_rate: number
  discount_allowed: boolean
  serial_tracking: boolean
  batch_tracking: boolean
  expiry_tracking: boolean
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
}

interface CartItem {
  product: Product
  quantity: number
  subtotal: number
  discount_amount: number
  tax_amount: number
  total: number
  notes?: string
}

interface Customer {
  id: string
  name: string
  email: string
  phone: string
  address: string
  loyalty_points: number
  loyalty_tier: 'bronze' | 'silver' | 'gold' | 'platinum'
  credit_limit: number
  payment_terms: string
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
}

interface PaymentMethod {
  id: string
  name: string
  type: 'cash' | 'card' | 'check' | 'gift_card' | 'mobile_payment' | 'crypto'
  icon: string
  processing_fee: number
  is_active: boolean
}

interface POSTransaction {
  id: string
  transaction_number: string
  customer_id?: string
  customer_name?: string
  cashier_id: string
  cashier_name: string
  register_id: string
  register_name: string
  session_id: string
  subtotal: number
  discount_amount: number
  tax_amount: number
  tip_amount: number
  total_amount: number
  payment_method: string
  payment_reference: string
  status: 'pending' | 'completed' | 'cancelled' | 'refunded'
  transaction_date: string
  items: CartItem[]
  notes?: string
  created_at: string
  updated_at: string
}

interface POSSession {
  id: string
  session_number: string
  cashier_id: string
  cashier_name: string
  register_id: string
  register_name: string
  opening_cash: number
  closing_cash?: number
  total_sales: number
  total_transactions: number
  status: 'open' | 'closed'
  opened_at: string
  closed_at?: string
  created_at: string
  updated_at: string
}

interface POSRegister {
  id: string
  name: string
  location: string
  ip_address: string
  status: 'online' | 'offline' | 'maintenance'
  last_activity: string
  current_session_id?: string
  current_cashier?: string
  created_at: string
  updated_at: string
}

interface POSAnalytics {
  today_sales: number
  today_transactions: number
  today_customers: number
  average_transaction_value: number
  top_products: Array<{
    product_id: string
    product_name: string
    quantity_sold: number
    revenue: number
  }>
  payment_methods: Array<{
    method: string
    count: number
    amount: number
    percentage: number
  }>
  hourly_sales: Array<{
    hour: number
    sales: number
    transactions: number
  }>
  customer_segments: Array<{
    segment: string
    count: number
    revenue: number
  }>
}

interface PWASettings {
  enable_offline_mode: boolean
  sync_frequency: number
  cache_size: number
  push_notifications: boolean
  background_sync: boolean
  auto_update: boolean
}

export function AdvancedPOSTerminal() {
  const [cart, setCart] = useState<CartItem[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null)
  const [paymentMethod, setPaymentMethod] = useState('cash')
  const [discountAmount, setDiscountAmount] = useState(0)
  const [discountType, setDiscountType] = useState<'percentage' | 'fixed'>('percentage')
  const [taxRate, setTaxRate] = useState(0.08)
  const [tipAmount, setTipAmount] = useState(0)
  const [notes, setNotes] = useState('')
  const [showCustomerSearch, setShowCustomerSearch] = useState(false)
  const [showPaymentModal, setShowPaymentModal] = useState(false)
  const [showReceiptModal, setShowReceiptModal] = useState(false)
  const [showScanner, setShowScanner] = useState(false)
  const [showAnalytics, setShowAnalytics] = useState(false)
  const [isOffline, setIsOffline] = useState(false)
  const [pwaSettings, setPwaSettings] = useState<PWASettings>({
    enable_offline_mode: true,
    sync_frequency: 30000,
    cache_size: 100,
    push_notifications: true,
    background_sync: true,
    auto_update: true
  })
  const [currentSession, setCurrentSession] = useState<POSSession | null>(null)
  const [currentRegister, setCurrentRegister] = useState<POSRegister | null>(null)
  const queryClient = useQueryClient()
  const scannerRef = useRef<HTMLVideoElement>(null)
  const streamRef = useRef<MediaStream | null>(null)

  // PWA Installation
  const [deferredPrompt, setDeferredPrompt] = useState<any>(null)
  const [isInstalled, setIsInstalled] = useState(false)

  useEffect(() => {
    // Check if app is already installed
    if (window.matchMedia('(display-mode: standalone)').matches) {
      setIsInstalled(true)
    }

    // Listen for beforeinstallprompt event
    window.addEventListener('beforeinstallprompt', (e) => {
      e.preventDefault()
      setDeferredPrompt(e)
    })

    // Listen for appinstalled event
    window.addEventListener('appinstalled', () => {
      setIsInstalled(true)
      setDeferredPrompt(null)
    })

    // Check online/offline status
    const handleOnline = () => setIsOffline(false)
    const handleOffline = () => setIsOffline(true)

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    // Initialize PWA features
    initializePWA()

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  const initializePWA = async () => {
    // Register service worker
    if ('serviceWorker' in navigator) {
      try {
        const registration = await navigator.serviceWorker.register('/sw.js')
        console.log('Service Worker registered:', registration)
      } catch (error) {
        console.log('Service Worker registration failed:', error)
      }
    }

    // Request notification permission
    if ('Notification' in window && Notification.permission === 'default') {
      await Notification.requestPermission()
    }

    // Initialize camera for barcode scanning
    if (navigator.mediaDevices && navigator.mediaDevices.getUserMedia) {
      // Camera will be initialized when scanner is opened
    }
  }

  const installPWA = async () => {
    if (deferredPrompt) {
      deferredPrompt.prompt()
      const { outcome } = await deferredPrompt.userChoice
      if (outcome === 'accepted') {
        console.log('PWA installed')
      }
      setDeferredPrompt(null)
    }
  }

  // Fetch products
  const { data: productsData } = useQuery({
    queryKey: ['pos-products', searchQuery],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (searchQuery) params.append('search', searchQuery)
      
      const response = await api.get(`/inventory/products?${params}`)
      return response.data.data.products as Product[]
    },
    enabled: !isOffline || pwaSettings.enable_offline_mode,
  })

  // Fetch customers
  const { data: customersData } = useQuery({
    queryKey: ['pos-customers'],
    queryFn: async () => {
      const response = await api.get('/customers')
      return response.data.data.customers as Customer[]
    },
    enabled: !isOffline || pwaSettings.enable_offline_mode,
  })

  // Fetch payment methods
  const { data: paymentMethodsData } = useQuery({
    queryKey: ['pos-payment-methods'],
    queryFn: async () => {
      const response = await api.get('/pos/payment-methods')
      return response.data.data as PaymentMethod[]
    },
  })

  // Fetch current session
  const { data: sessionData } = useQuery({
    queryKey: ['pos-current-session'],
    queryFn: async () => {
      const response = await api.get('/pos/sessions/current')
      return response.data.data as POSSession
    },
  })

  // Fetch current register
  const { data: registerData } = useQuery({
    queryKey: ['pos-current-register'],
    queryFn: async () => {
      const response = await api.get('/pos/registers/current')
      return response.data.data as POSRegister
    },
  })

  // Fetch analytics
  const { data: analyticsData } = useQuery({
    queryKey: ['pos-analytics'],
    queryFn: async () => {
      const response = await api.get('/pos/analytics')
      return response.data.data as POSAnalytics
    },
  })

  // Create transaction mutation
  const createTransaction = useMutation({
    mutationFn: async (transactionData: Partial<POSTransaction>) => {
      const response = await api.post('/pos/transactions', transactionData)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pos-transactions'] })
      setCart([])
      setDiscountAmount(0)
      setTipAmount(0)
      setNotes('')
      setSelectedCustomer(null)
      setShowPaymentModal(false)
      setShowReceiptModal(true)
    },
  })

  // Start scanner
  const startScanner = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ 
        video: { facingMode: 'environment' } 
      })
      streamRef.current = stream
      if (scannerRef.current) {
        scannerRef.current.srcObject = stream
      }
      setShowScanner(true)
    } catch (error) {
      console.error('Error accessing camera:', error)
      alert('Unable to access camera for barcode scanning')
    }
  }

  // Stop scanner
  const stopScanner = () => {
    if (streamRef.current) {
      streamRef.current.getTracks().forEach(track => track.stop())
      streamRef.current = null
    }
    setShowScanner(false)
  }

  // Add to cart
  const addToCart = (product: Product) => {
    const existingItem = cart.find(item => item.product.id === product.id)
    
    if (existingItem) {
      setCart(cart.map(item =>
        item.product.id === product.id
          ? { ...item, quantity: item.quantity + 1, subtotal: (item.quantity + 1) * product.price }
          : item
      ))
    } else {
      setCart([...cart, {
        product,
        quantity: 1,
        subtotal: product.price,
        discount_amount: 0,
        tax_amount: product.price * product.tax_rate,
        total: product.price + (product.price * product.tax_rate)
      }])
    }
  }

  // Remove from cart
  const removeFromCart = (productId: string) => {
    setCart(cart.filter(item => item.product.id !== productId))
  }

  // Update quantity
  const updateQuantity = (productId: string, quantity: number) => {
    if (quantity <= 0) {
      removeFromCart(productId)
      return
    }
    
    setCart(cart.map(item =>
      item.product.id === productId
        ? { ...item, quantity, subtotal: quantity * item.product.price }
        : item
    ))
  }

  // Calculate totals
  const calculateTotals = () => {
    const subtotal = cart.reduce((sum, item) => sum + item.subtotal, 0)
    const discount = discountType === 'percentage' 
      ? (subtotal * discountAmount) / 100 
      : discountAmount
    const taxableAmount = subtotal - discount
    const tax = taxableAmount * taxRate
    const total = taxableAmount + tax + tipAmount

    return { subtotal, discount, tax, tip: tipAmount, total }
  }

  // Handle checkout
  const handleCheckout = () => {
    if (cart.length === 0) {
      alert('Cart is empty')
      return
    }

    setShowPaymentModal(true)
  }

  // Process payment
  const processPayment = () => {
    const { subtotal, discount, tax, tip, total } = calculateTotals()
    
    const transactionData = {
      customer_id: selectedCustomer?.id || null,
      subtotal,
      discount_amount: discount,
      tax_amount: tax,
      tip_amount: tip,
      total_amount: total,
      payment_method: paymentMethod,
      payment_reference: `REF-${Date.now()}`,
      status: 'completed',
      transaction_date: new Date().toISOString(),
      items: cart,
      notes: notes,
    }

    createTransaction.mutate(transactionData)
  }

  // Print receipt
  const printReceipt = () => {
    if ('print' in window) {
      window.print()
    } else {
      // Fallback for mobile devices
      const receiptContent = generateReceiptContent()
      const printWindow = window.open('', '_blank')
      printWindow?.document.write(receiptContent)
      printWindow?.print()
    }
  }

  // Generate receipt content
  const generateReceiptContent = () => {
    const { subtotal, discount, tax, tip, total } = calculateTotals()
    const currentDate = new Date().toLocaleString()
    
    return `
      <html>
        <head>
          <title>Receipt</title>
          <style>
            body { font-family: monospace; font-size: 12px; margin: 0; padding: 20px; }
            .header { text-align: center; margin-bottom: 20px; }
            .item { display: flex; justify-content: space-between; margin-bottom: 5px; }
            .total { border-top: 1px solid #000; padding-top: 10px; margin-top: 10px; }
            .footer { text-align: center; margin-top: 20px; font-size: 10px; }
          </style>
        </head>
        <body>
          <div class="header">
            <h2>RECEIPT</h2>
            <p>Date: ${currentDate}</p>
            <p>Transaction: ${createTransaction.data?.transaction_number || 'N/A'}</p>
          </div>
          
          <div class="items">
            ${cart.map(item => `
              <div class="item">
                <span>${item.product.name} x${item.quantity}</span>
                <span>$${item.subtotal.toFixed(2)}</span>
              </div>
            `).join('')}
          </div>
          
          <div class="total">
            <div class="item">
              <span>Subtotal:</span>
              <span>$${subtotal.toFixed(2)}</span>
            </div>
            <div class="item">
              <span>Discount:</span>
              <span>-$${discount.toFixed(2)}</span>
            </div>
            <div class="item">
              <span>Tax:</span>
              <span>$${tax.toFixed(2)}</span>
            </div>
            <div class="item">
              <span>Tip:</span>
              <span>$${tip.toFixed(2)}</span>
            </div>
            <div class="item" style="font-weight: bold;">
              <span>TOTAL:</span>
              <span>$${total.toFixed(2)}</span>
            </div>
          </div>
          
          <div class="footer">
            <p>Thank you for your business!</p>
            <p>Payment Method: ${paymentMethod.toUpperCase()}</p>
          </div>
        </body>
      </html>
    `
  }

  // Send receipt via email/SMS
  const sendReceipt = async (method: 'email' | 'sms') => {
    if (!selectedCustomer) {
      alert('Please select a customer to send receipt')
      return
    }

    try {
      await api.post('/pos/receipts/send', {
        customer_id: selectedCustomer.id,
        method: method,
        transaction_id: createTransaction.data?.id
      })
      alert(`Receipt sent via ${method.toUpperCase()}`)
    } catch (error) {
      console.error('Error sending receipt:', error)
      alert('Failed to send receipt')
    }
  }

  const { subtotal, discount, tax, tip, total } = calculateTotals()
  const products = productsData || []
  const customers = customersData || []
  const paymentMethods = paymentMethodsData || []
  const analytics = analyticsData

  return (
    <div className="h-screen bg-gray-100 flex flex-col">
      {/* PWA Header */}
      <div className="bg-blue-600 text-white p-2 flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <ShoppingCart className="h-5 w-5" />
          <span className="font-semibold">POS Terminal</span>
          {isOffline && (
            <span className="bg-orange-500 px-2 py-1 rounded text-xs">OFFLINE</span>
          )}
        </div>
        <div className="flex items-center space-x-2">
          {!isInstalled && deferredPrompt && (
            <button
              onClick={installPWA}
              className="bg-white text-blue-600 px-3 py-1 rounded text-sm font-medium"
            >
              Install App
            </button>
          )}
          <button
            onClick={() => setShowAnalytics(true)}
            className="p-1 hover:bg-blue-700 rounded"
          >
            <BarChart3 className="h-4 w-4" />
          </button>
          <button className="p-1 hover:bg-blue-700 rounded">
            <Settings className="h-4 w-4" />
          </button>
        </div>
      </div>

      <div className="flex-1 flex">
        {/* Left Panel - Products */}
        <div className="w-2/3 p-4">
          <div className="bg-white rounded-lg shadow-sm h-full flex flex-col">
            {/* Search and Scanner */}
            <div className="p-4 border-b">
              <div className="flex items-center space-x-2">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                  <input
                    type="text"
                    placeholder="Search products or scan barcode..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
                <button
                  onClick={startScanner}
                  className="flex items-center px-3 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-lg hover:bg-blue-700"
                >
                  <Barcode className="h-4 w-4 mr-2" />
                  Scan
                </button>
              </div>
            </div>

            {/* Products Grid */}
            <div className="flex-1 p-4 overflow-y-auto">
              <div className="grid grid-cols-4 gap-4">
                {products.map(product => (
                  <div
                    key={product.id}
                    onClick={() => addToCart(product)}
                    className="bg-gray-50 rounded-lg p-4 cursor-pointer hover:bg-gray-100 border border-gray-200 transition-colors"
                  >
                    {product.image_url && (
                      <img
                        src={product.image_url}
                        alt={product.name}
                        className="w-full h-20 object-cover rounded mb-2"
                      />
                    )}
                    <div className="text-sm font-medium text-gray-900 mb-1 truncate">
                      {product.name}
                    </div>
                    <div className="text-xs text-gray-500 mb-2">
                      {product.sku}
                    </div>
                    <div className="text-lg font-bold text-green-600">
                      ${product.price.toFixed(2)}
                    </div>
                    {product.stock_quantity <= product.min_stock_level && (
                      <div className="text-xs text-red-600 mt-1">
                        Low Stock: {product.stock_quantity}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Right Panel - Cart & Checkout */}
        <div className="w-1/3 p-4">
          <div className="bg-white rounded-lg shadow-sm h-full flex flex-col">
            {/* Cart Header */}
            <div className="p-4 border-b">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-gray-900">Shopping Cart</h2>
                <span className="text-sm text-gray-500">{cart.length} items</span>
              </div>
            </div>

            {/* Cart Items */}
            <div className="flex-1 p-4 overflow-y-auto">
              {cart.length === 0 ? (
                <div className="text-center text-gray-500 py-8">
                  <ShoppingCart className="h-12 w-12 mx-auto mb-2 text-gray-300" />
                  <p>Cart is empty</p>
                  <p className="text-sm">Add products to get started</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {cart.map(item => (
                    <div key={item.product.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                      <div className="flex-1">
                        <div className="font-medium text-gray-900">{item.product.name}</div>
                        <div className="text-sm text-gray-500">${item.product.price.toFixed(2)} each</div>
                      </div>
                      <div className="flex items-center space-x-2">
                        <button
                          onClick={() => updateQuantity(item.product.id, item.quantity - 1)}
                          className="p-1 text-gray-500 hover:text-gray-700"
                        >
                          <Minus className="h-4 w-4" />
                        </button>
                        <span className="w-8 text-center font-medium">{item.quantity}</span>
                        <button
                          onClick={() => updateQuantity(item.product.id, item.quantity + 1)}
                          className="p-1 text-gray-500 hover:text-gray-700"
                        >
                          <Plus className="h-4 w-4" />
                        </button>
                        <button
                          onClick={() => removeFromCart(item.product.id)}
                          className="p-1 text-red-500 hover:text-red-700 ml-2"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                      <div className="font-medium text-gray-900 ml-4">
                        ${item.subtotal.toFixed(2)}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Customer Selection */}
            <div className="p-4 border-t">
              <div className="flex items-center justify-between mb-2">
                <label className="text-sm font-medium text-gray-700">Customer</label>
                <button
                  onClick={() => setShowCustomerSearch(true)}
                  className="text-blue-600 hover:text-blue-800 text-sm"
                >
                  {selectedCustomer ? 'Change' : 'Select'}
                </button>
              </div>
              {selectedCustomer ? (
                <div className="flex items-center justify-between p-2 bg-blue-50 rounded">
                  <div>
                    <div className="font-medium text-gray-900">{selectedCustomer.name}</div>
                    <div className="text-sm text-gray-500">{selectedCustomer.email}</div>
                  </div>
                  <button
                    onClick={() => setSelectedCustomer(null)}
                    className="text-red-500 hover:text-red-700"
                  >
                    <X className="h-4 w-4" />
                  </button>
                </div>
              ) : (
                <div className="text-sm text-gray-500">No customer selected</div>
              )}
            </div>

            {/* Totals */}
            <div className="p-4 border-t">
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-gray-600">Subtotal:</span>
                  <span className="font-medium">${subtotal.toFixed(2)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Discount:</span>
                  <span className="font-medium text-red-600">-${discount.toFixed(2)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Tax:</span>
                  <span className="font-medium">${tax.toFixed(2)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Tip:</span>
                  <span className="font-medium">${tip.toFixed(2)}</span>
                </div>
                <div className="flex justify-between text-lg font-bold border-t pt-2">
                  <span>Total:</span>
                  <span className="text-green-600">${total.toFixed(2)}</span>
                </div>
              </div>
            </div>

            {/* Checkout Button */}
            <div className="p-4 border-t">
              <button
                onClick={handleCheckout}
                disabled={cart.length === 0}
                className="w-full flex items-center justify-center px-4 py-3 text-sm font-medium text-white bg-green-600 border border-transparent rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <CreditCard className="h-4 w-4 mr-2" />
                Checkout - ${total.toFixed(2)}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Barcode Scanner Modal */}
      {showScanner && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Barcode Scanner</h3>
              <button
                onClick={stopScanner}
                className="text-gray-500 hover:text-gray-700"
              >
                <X className="h-6 w-6" />
              </button>
            </div>
            
            <div className="text-center">
              <div className="bg-gray-100 rounded-lg p-8 mb-4">
                <video
                  ref={scannerRef}
                  autoPlay
                  playsInline
                  className="w-full h-48 object-cover rounded"
                />
                <p className="text-gray-600 mt-2">Position barcode in front of camera</p>
              </div>
              
              <div className="space-y-2">
                <button className="w-full px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-lg hover:bg-blue-700">
                  Start Scanning
                </button>
                <button className="w-full px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50">
                  Manual Entry
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Customer Search Modal */}
      {showCustomerSearch && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Select Customer</h3>
              <button
                onClick={() => setShowCustomerSearch(false)}
                className="text-gray-500 hover:text-gray-700"
              >
                <X className="h-6 w-6" />
              </button>
            </div>
            
            <div className="space-y-2">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search customers..."
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
              
              <div className="max-h-64 overflow-y-auto">
                {customers.map(customer => (
                  <div
                    key={customer.id}
                    onClick={() => {
                      setSelectedCustomer(customer)
                      setShowCustomerSearch(false)
                    }}
                    className="p-3 hover:bg-gray-50 cursor-pointer border-b"
                  >
                    <div className="font-medium text-gray-900">{customer.name}</div>
                    <div className="text-sm text-gray-500">{customer.email}</div>
                    <div className="text-xs text-gray-400">
                      {customer.phone} • {customer.loyalty_points} points
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Payment Modal */}
      {showPaymentModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Payment</h3>
              <button
                onClick={() => setShowPaymentModal(false)}
                className="text-gray-500 hover:text-gray-700"
              >
                <X className="h-6 w-6" />
              </button>
            </div>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Total Amount</label>
                <div className="text-2xl font-bold text-green-600">${total.toFixed(2)}</div>
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Payment Method</label>
                <select
                  value={paymentMethod}
                  onChange={(e) => setPaymentMethod(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  {paymentMethods.map(method => (
                    <option key={method.id} value={method.type}>{method.name}</option>
                  ))}
                </select>
              </div>

              {paymentMethod === 'cash' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Amount Received</label>
                  <input
                    type="number"
                    step="0.01"
                    min={total}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
              )}

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Notes (Optional)</label>
                <textarea
                  rows={2}
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>
            
            <div className="flex justify-end space-x-2 mt-6">
              <button
                onClick={() => setShowPaymentModal(false)}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={processPayment}
                disabled={createTransaction.isPending}
                className="px-4 py-2 text-sm font-medium text-white bg-green-600 border border-transparent rounded-lg hover:bg-green-700 disabled:opacity-50"
              >
                {createTransaction.isPending ? 'Processing...' : 'Process Payment'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Receipt Modal */}
      {showReceiptModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Transaction Complete</h3>
              <button
                onClick={() => setShowReceiptModal(false)}
                className="text-gray-500 hover:text-gray-700"
              >
                <X className="h-6 w-6" />
              </button>
            </div>
            
            <div className="text-center mb-6">
              <CheckCircle className="h-16 w-16 text-green-600 mx-auto mb-2" />
              <p className="text-lg font-medium text-gray-900">Payment Successful!</p>
              <p className="text-sm text-gray-500">Transaction #{createTransaction.data?.transaction_number}</p>
            </div>
            
            <div className="space-y-2">
              <button
                onClick={printReceipt}
                className="w-full flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-lg hover:bg-blue-700"
              >
                <Printer className="h-4 w-4 mr-2" />
                Print Receipt
              </button>
              
              {selectedCustomer && (
                <>
                  <button
                    onClick={() => sendReceipt('email')}
                    className="w-full flex items-center justify-center px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
                  >
                    <Mail className="h-4 w-4 mr-2" />
                    Email Receipt
                  </button>
                  
                  <button
                    onClick={() => sendReceipt('sms')}
                    className="w-full flex items-center justify-center px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
                  >
                    <Phone className="h-4 w-4 mr-2" />
                    SMS Receipt
                  </button>
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Analytics Modal */}
      {showAnalytics && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">POS Analytics</h3>
              <button
                onClick={() => setShowAnalytics(false)}
                className="text-gray-500 hover:text-gray-700"
              >
                <X className="h-6 w-6" />
              </button>
            </div>
            
            <div className="grid grid-cols-2 gap-4 mb-6">
              <div className="bg-blue-50 p-4 rounded-lg">
                <div className="text-2xl font-bold text-blue-600">${analytics?.today_sales?.toFixed(2) || 0}</div>
                <div className="text-sm text-gray-600">Today's Sales</div>
              </div>
              <div className="bg-green-50 p-4 rounded-lg">
                <div className="text-2xl font-bold text-green-600">{analytics?.today_transactions || 0}</div>
                <div className="text-sm text-gray-600">Today's Transactions</div>
              </div>
              <div className="bg-purple-50 p-4 rounded-lg">
                <div className="text-2xl font-bold text-purple-600">{analytics?.today_customers || 0}</div>
                <div className="text-sm text-gray-600">Today's Customers</div>
              </div>
              <div className="bg-orange-50 p-4 rounded-lg">
                <div className="text-2xl font-bold text-orange-600">${analytics?.average_transaction_value?.toFixed(2) || 0}</div>
                <div className="text-sm text-gray-600">Avg Transaction</div>
              </div>
            </div>

            <div className="space-y-4">
              <div>
                <h4 className="font-medium text-gray-900 mb-2">Top Products</h4>
                <div className="space-y-2">
                  {analytics?.top_products?.slice(0, 5).map((product, index) => (
                    <div key={product.product_id} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                      <span className="text-sm">{index + 1}. {product.product_name}</span>
                      <span className="text-sm font-medium">${product.revenue.toFixed(2)}</span>
                    </div>
                  ))}
                </div>
              </div>

              <div>
                <h4 className="font-medium text-gray-900 mb-2">Payment Methods</h4>
                <div className="space-y-2">
                  {analytics?.payment_methods?.map((method, index) => (
                    <div key={index} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                      <span className="text-sm">{method.method}</span>
                      <span className="text-sm font-medium">${method.amount.toFixed(2)} ({method.percentage.toFixed(1)}%)</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
