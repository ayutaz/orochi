import { useCallback, useEffect, useRef, useState } from 'react';

interface WebSocketOptions {
  reconnectDelay?: number;
  maxReconnectAttempts?: number;
  messageQueueSize?: number;
  batchInterval?: number;
}

interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: number;
}

export const useWebSocketOptimized = (url: string, options: WebSocketOptions = {}) => {
  const {
    reconnectDelay = 3000,
    maxReconnectAttempts = 10,
    messageQueueSize = 100,
    batchInterval = 100,
  } = options;

  const [isConnected, setIsConnected] = useState(false);
  const [messageQueue, setMessageQueue] = useState<WebSocketMessage[]>([]);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const messageBufferRef = useRef<WebSocketMessage[]>([]);
  const batchTimerRef = useRef<NodeJS.Timeout | null>(null);

  // Process batched messages
  const processBatchedMessages = useCallback(() => {
    if (messageBufferRef.current.length > 0) {
      setMessageQueue((prev) => {
        const newQueue = [...prev, ...messageBufferRef.current];
        // Keep only the last N messages
        return newQueue.slice(-messageQueueSize);
      });
      messageBufferRef.current = [];
    }
  }, [messageQueueSize]);

  // Send message with queuing
  const sendMessage = useCallback((message: any) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    }
  }, []);

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    const ws = new WebSocket(url);

    ws.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
      reconnectAttemptsRef.current = 0;
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        const message: WebSocketMessage = {
          type: data.type,
          data: data.data,
          timestamp: Date.now(),
        };

        // Add to buffer
        messageBufferRef.current.push(message);

        // Clear existing timer and set new one
        if (batchTimerRef.current) {
          clearTimeout(batchTimerRef.current);
        }

        batchTimerRef.current = setTimeout(() => {
          processBatchedMessages();
        }, batchInterval);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      setIsConnected(false);
      wsRef.current = null;

      // Reconnect logic
      if (reconnectAttemptsRef.current < maxReconnectAttempts) {
        reconnectAttemptsRef.current++;
        console.log(`Reconnecting... Attempt ${reconnectAttemptsRef.current}`);
        setTimeout(connect, reconnectDelay);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    wsRef.current = ws;
  }, [url, reconnectDelay, maxReconnectAttempts, batchInterval, processBatchedMessages]);

  // Disconnect
  const disconnect = useCallback(() => {
    if (batchTimerRef.current) {
      clearTimeout(batchTimerRef.current);
      processBatchedMessages();
    }
    
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, [processBatchedMessages]);

  // Get messages by type
  const getMessagesByType = useCallback((type: string) => {
    return messageQueue.filter((msg) => msg.type === type);
  }, [messageQueue]);

  // Get latest message by type
  const getLatestMessageByType = useCallback((type: string) => {
    const messages = getMessagesByType(type);
    return messages[messages.length - 1] || null;
  }, [getMessagesByType]);

  useEffect(() => {
    connect();
    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    isConnected,
    sendMessage,
    messageQueue,
    getMessagesByType,
    getLatestMessageByType,
    reconnect: connect,
    disconnect,
  };
};