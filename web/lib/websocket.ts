import { io, Socket } from 'socket.io-client'
import { Message, Space } from './api'

// WebSocket events
export enum WebSocketEvent {
  CONNECT = 'connect',
  DISCONNECT = 'disconnect',
  ERROR = 'error',
  JOIN_SPACE = 'join_space',
  LEAVE_SPACE = 'leave_space',
  NEW_MESSAGE = 'new_message',
  SPACE_UPDATED = 'space_updated',
  USER_JOINED = 'user_joined',
  USER_LEFT = 'user_left',
  TYPING = 'typing',
  STOP_TYPING = 'stop_typing',
}

// WebSocket message types
export interface WebSocketMessage {
  type: string
  payload: any
}

// Typing indicator
export interface TypingIndicator {
  spaceId: string
  userId: string
  username: string
  isTyping: boolean
}

// Class to manage WebSocket connections
class WebSocketService {
  private socket: Socket | null = null
  private messageListeners: Map<string, Set<(message: Message) => void>> = new Map()
  private spaceUpdateListeners: Map<string, Set<(space: Space) => void>> = new Map()
  private typingListeners: Map<string, Set<(data: TypingIndicator) => void>> = new Map()
  private userActivityListeners: Map<string, Set<(data: { userId: string, username: string }) => void>> = new Map()
  
  // Initialize the WebSocket connection
  connect(): void {
    if (this.socket) return
    
    const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws'
    
    // Get anonymous user info
    let anonymousUser = null
    if (typeof window !== 'undefined') {
      const storedUser = localStorage.getItem('anonymous_user')
      if (storedUser) {
        try {
          anonymousUser = JSON.parse(storedUser)
        } catch (e) {
          // Ignore parsing errors
        }
      }
    }
    
    this.socket = io(WS_URL, {
      transports: ['websocket'],
      auth: anonymousUser ? {
        anonymousId: anonymousUser.id,
        displayName: anonymousUser.displayName,
        color: anonymousUser.color
      } : undefined,
      reconnection: true,
      reconnectionAttempts: 5,
      reconnectionDelay: 1000,
    })
    
    // Set up event listeners
    this.socket.on(WebSocketEvent.CONNECT, () => {
      console.log('WebSocket connected')
    })
    
    this.socket.on(WebSocketEvent.DISCONNECT, () => {
      console.log('WebSocket disconnected')
    })
    
    this.socket.on(WebSocketEvent.ERROR, (error) => {
      console.error('WebSocket error:', error)
    })
    
    // Handle new messages
    this.socket.on(WebSocketEvent.NEW_MESSAGE, (message: Message) => {
      const listeners = this.messageListeners.get(message.spaceId)
      if (listeners) {
        listeners.forEach(listener => listener(message))
      }
    })
    
    // Handle space updates
    this.socket.on(WebSocketEvent.SPACE_UPDATED, (space: Space) => {
      const listeners = this.spaceUpdateListeners.get(space.id)
      if (listeners) {
        listeners.forEach(listener => listener(space))
      }
    })
    
    // Handle typing indicators
    this.socket.on(WebSocketEvent.TYPING, (data: TypingIndicator) => {
      const listeners = this.typingListeners.get(data.spaceId)
      if (listeners) {
        listeners.forEach(listener => listener({ ...data, isTyping: true }))
      }
    })
    
    this.socket.on(WebSocketEvent.STOP_TYPING, (data: TypingIndicator) => {
      const listeners = this.typingListeners.get(data.spaceId)
      if (listeners) {
        listeners.forEach(listener => listener({ ...data, isTyping: false }))
      }
    })
    
    // Handle user activity
    this.socket.on(WebSocketEvent.USER_JOINED, (data: { spaceId: string, userId: string, username: string }) => {
      const listeners = this.userActivityListeners.get(data.spaceId)
      if (listeners) {
        listeners.forEach(listener => listener({ userId: data.userId, username: data.username }))
      }
    })
    
    this.socket.on(WebSocketEvent.USER_LEFT, (data: { spaceId: string, userId: string, username: string }) => {
      const listeners = this.userActivityListeners.get(data.spaceId)
      if (listeners) {
        listeners.forEach(listener => listener({ userId: data.userId, username: data.username }))
      }
    })
  }
  
  // Disconnect the WebSocket
  disconnect(): void {
    if (this.socket) {
      this.socket.disconnect()
      this.socket = null
    }
  }
  
  // Join a space
  joinSpace(spaceId: string): void {
    if (!this.socket) this.connect()
    this.socket?.emit(WebSocketEvent.JOIN_SPACE, { spaceId })
  }
  
  // Leave a space
  leaveSpace(spaceId: string): void {
    this.socket?.emit(WebSocketEvent.LEAVE_SPACE, { spaceId })
  }
  
  // Send a typing indicator
  sendTypingIndicator(spaceId: string): void {
    this.socket?.emit(WebSocketEvent.TYPING, { spaceId })
  }
  
  // Send a stop typing indicator
  sendStopTypingIndicator(spaceId: string): void {
    this.socket?.emit(WebSocketEvent.STOP_TYPING, { spaceId })
  }
  
  // Subscribe to new messages in a space
  subscribeToMessages(spaceId: string, callback: (message: Message) => void): () => void {
    if (!this.messageListeners.has(spaceId)) {
      this.messageListeners.set(spaceId, new Set())
    }
    
    this.messageListeners.get(spaceId)?.add(callback)
    
    // Return unsubscribe function
    return () => {
      const listeners = this.messageListeners.get(spaceId)
      if (listeners) {
        listeners.delete(callback)
        if (listeners.size === 0) {
          this.messageListeners.delete(spaceId)
        }
      }
    }
  }
  
  // Subscribe to space updates
  subscribeToSpaceUpdates(spaceId: string, callback: (space: Space) => void): () => void {
    if (!this.spaceUpdateListeners.has(spaceId)) {
      this.spaceUpdateListeners.set(spaceId, new Set())
    }
    
    this.spaceUpdateListeners.get(spaceId)?.add(callback)
    
    // Return unsubscribe function
    return () => {
      const listeners = this.spaceUpdateListeners.get(spaceId)
      if (listeners) {
        listeners.delete(callback)
        if (listeners.size === 0) {
          this.spaceUpdateListeners.delete(spaceId)
        }
      }
    }
  }
  
  // Subscribe to typing indicators
  subscribeToTypingIndicators(spaceId: string, callback: (data: TypingIndicator) => void): () => void {
    if (!this.typingListeners.has(spaceId)) {
      this.typingListeners.set(spaceId, new Set())
    }
    
    this.typingListeners.get(spaceId)?.add(callback)
    
    // Return unsubscribe function
    return () => {
      const listeners = this.typingListeners.get(spaceId)
      if (listeners) {
        listeners.delete(callback)
        if (listeners.size === 0) {
          this.typingListeners.delete(spaceId)
        }
      }
    }
  }
  
  // Subscribe to user activity (join/leave)
  subscribeToUserActivity(
    spaceId: string, 
    callback: (data: { userId: string, username: string }) => void
  ): () => void {
    if (!this.userActivityListeners.has(spaceId)) {
      this.userActivityListeners.set(spaceId, new Set())
    }
    
    this.userActivityListeners.get(spaceId)?.add(callback)
    
    // Return unsubscribe function
    return () => {
      const listeners = this.userActivityListeners.get(spaceId)
      if (listeners) {
        listeners.delete(callback)
        if (listeners.size === 0) {
          this.userActivityListeners.delete(spaceId)
        }
      }
    }
  }
}

// Create a singleton instance
export const webSocketService = new WebSocketService()

// Hook to use WebSocket in components
export function useWebSocket() {
  return webSocketService
} 