import React, { createContext, useContext, useEffect, useState } from 'react'

interface WebSocketContextType {
  isConnected: boolean
  lastMessage: any
}

const WebSocketContext = createContext<WebSocketContextType>({
  isConnected: false,
  lastMessage: null,
})

export const useWebSocket = () => useContext(WebSocketContext)

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [isConnected, setIsConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState<any>(null)
  const [ws, setWs] = useState<WebSocket | null>(null)

  useEffect(() => {
    const connectWebSocket = () => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/ws`
      
      const websocket = new WebSocket(wsUrl)

      websocket.onopen = () => {
        console.log('WebSocket connected')
        setIsConnected(true)
      }

      websocket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          setLastMessage(data)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      websocket.onclose = () => {
        console.log('WebSocket disconnected')
        setIsConnected(false)
        // Reconnect after 3 seconds
        setTimeout(connectWebSocket, 3000)
      }

      websocket.onerror = (error) => {
        console.error('WebSocket error:', error)
      }

      setWs(websocket)
    }

    connectWebSocket()

    return () => {
      if (ws) {
        ws.close()
      }
    }
  }, [])

  return (
    <WebSocketContext.Provider value={{ isConnected, lastMessage }}>
      {children}
    </WebSocketContext.Provider>
  )
}