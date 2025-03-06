'use client'

import { useEffect } from 'react'
import { Provider } from 'react-redux'
import { store } from '@/lib/store'
import { webSocketService } from '@/lib/websocket'
import { ThemeProvider } from '@/components/ThemeProvider'

// Component to initialize WebSocket connection
function WebSocketInitializer() {
  useEffect(() => {
    // Connect to WebSocket immediately - no authentication required
    webSocketService.connect()
    
    // Cleanup on unmount
    return () => {
      webSocketService.disconnect()
    }
  }, [])
  
  return null
}

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <Provider store={store}>
      <ThemeProvider>
        <WebSocketInitializer />
        {children}
      </ThemeProvider>
    </Provider>
  )
} 