'use client'

import { useEffect, useState, useRef } from 'react'
import { useParams } from 'next/navigation'
import { useDispatch, useSelector } from 'react-redux'
import { formatDistanceToNow } from 'date-fns'
import { api, Message } from '@/lib/api'
import { setActiveSpace, joinSpace } from '@/lib/features/spaces/spacesSlice'
import { RootState } from '@/lib/store'
import { useWebSocket, TypingIndicator } from '@/lib/websocket'

export default function SpaceDetailPage() {
  const params = useParams()
  const dispatch = useDispatch()
  const spaceId = params?.id as string
  const { joinedSpaces } = useSelector((state: RootState) => state.spaces)
  const { user } = useSelector((state: RootState) => state.user)
  const isJoined = joinedSpaces.includes(spaceId)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const webSocket = useWebSocket()
  
  // Local state for real-time messages
  const [localMessages, setLocalMessages] = useState<Message[]>([])
  const [typingUsers, setTypingUsers] = useState<{ [key: string]: string }>({})
  const [isTyping, setIsTyping] = useState(false)
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  
  // API queries
  const { data: space, isLoading: spaceLoading } = api.useGetSpaceByIdQuery(spaceId)
  const { data: initialMessages, isLoading: messagesLoading } = api.useGetSpaceMessagesQuery(spaceId, {
    skip: !isJoined
  })
  
  const [sendMessage, { isLoading: isSending }] = api.useSendMessageMutation()
  
  // Set active space and connect to WebSocket
  useEffect(() => {
    if (spaceId) {
      dispatch(setActiveSpace(spaceId))
      
      // Join the space if already in joinedSpaces
      if (isJoined) {
        webSocket.joinSpace(spaceId)
      }
    }
    
    return () => {
      dispatch(setActiveSpace(null))
      if (isJoined && spaceId) {
        webSocket.leaveSpace(spaceId)
      }
    }
  }, [spaceId, dispatch, webSocket, isJoined])
  
  // Initialize local messages from API data
  useEffect(() => {
    if (initialMessages) {
      setLocalMessages(initialMessages)
    }
  }, [initialMessages])
  
  // Subscribe to new messages
  useEffect(() => {
    if (!isJoined || !spaceId) return
    
    const unsubscribe = webSocket.subscribeToMessages(spaceId, (message) => {
      setLocalMessages(prev => [...prev, message])
    })
    
    return unsubscribe
  }, [webSocket, spaceId, isJoined])
  
  // Subscribe to typing indicators
  useEffect(() => {
    if (!isJoined || !spaceId) return
    
    const unsubscribe = webSocket.subscribeToTypingIndicators(spaceId, (data: TypingIndicator) => {
      if (data.userId === user?.id) return // Ignore own typing indicators
      
      if (data.isTyping) {
        setTypingUsers(prev => ({ ...prev, [data.userId]: data.username }))
      } else {
        setTypingUsers(prev => {
          const newState = { ...prev }
          delete newState[data.userId]
          return newState
        })
      }
    })
    
    return unsubscribe
  }, [webSocket, spaceId, isJoined, user?.id])
  
  // Scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [localMessages])
  
  const handleJoin = () => {
    dispatch(joinSpace(spaceId))
    webSocket.joinSpace(spaceId)
  }
  
  const handleSendMessage = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const form = e.currentTarget
    const formData = new FormData(form)
    const content = formData.get('message') as string
    
    if (content.trim() && spaceId) {
      // Clear typing indicator
      if (isTyping) {
        setIsTyping(false)
        webSocket.sendStopTypingIndicator(spaceId)
      }
      
      // Send message via API
      try {
        const result = await sendMessage({ spaceId, content }).unwrap()
        
        // Optimistically add message to local state
        // The WebSocket will deliver the official message to all clients
        form.reset()
      } catch (error) {
        console.error('Failed to send message:', error)
      }
    }
  }
  
  const handleTyping = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!isJoined || !spaceId) return
    
    // Send typing indicator if not already typing
    if (!isTyping && e.target.value.trim()) {
      setIsTyping(true)
      webSocket.sendTypingIndicator(spaceId)
    }
    
    // Clear previous timeout
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current)
    }
    
    // Set timeout to clear typing indicator
    typingTimeoutRef.current = setTimeout(() => {
      if (isTyping) {
        setIsTyping(false)
        webSocket.sendStopTypingIndicator(spaceId)
      }
    }, 3000)
  }
  
  if (spaceLoading) {
    return (
      <div className="container-wide py-12">
        <div className="animate-pulse">
          <div className="h-8 w-1/3 rounded bg-gray-200"></div>
          <div className="mt-4 h-4 w-2/3 rounded bg-gray-200"></div>
          <div className="mt-8 h-64 rounded-lg bg-gray-200"></div>
        </div>
      </div>
    )
  }
  
  if (!space) {
    return (
      <div className="container-narrow py-12">
        <div className="rounded-md bg-red-50 p-4">
          <div className="flex">
            <div className="text-sm text-red-700">
              <p>Space not found or has been dissolved.</p>
            </div>
          </div>
        </div>
      </div>
    )
  }
  
  return (
    <div className="container-wide py-12">
      <div className="mb-8">
        <div className="flex items-center gap-2">
          <h1 className="text-3xl font-bold text-gray-900">{space.title}</h1>
          <span className={`ml-2 rounded-full px-2.5 py-0.5 text-xs font-medium ${
            space.lifecycleStage === 'growing' ? 'bg-green-100 text-green-800' :
            space.lifecycleStage === 'peak' ? 'bg-blue-100 text-blue-800' :
            space.lifecycleStage === 'waning' ? 'bg-yellow-100 text-yellow-800' :
            space.lifecycleStage === 'dissolving' ? 'bg-red-100 text-red-800' :
            'bg-gray-100 text-gray-800'
          }`}>
            {space.lifecycleStage.charAt(0).toUpperCase() + space.lifecycleStage.slice(1)}
          </span>
        </div>
        <p className="mt-2 text-gray-600">{space.description}</p>
        
        <div className="mt-4 flex flex-wrap gap-2">
          {space.topicTags.map((tag: string) => (
            <span key={tag} className="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-800">
              {tag}
            </span>
          ))}
        </div>
        
        <div className="mt-4 flex items-center gap-4 text-sm text-gray-500">
          <span>{space.userCount} participants</span>
          <span>{space.messageCount} messages</span>
          <span>Created {formatDistanceToNow(new Date(space.createdAt), { addSuffix: true })}</span>
          <span>Last active {formatDistanceToNow(new Date(space.lastActive), { addSuffix: true })}</span>
        </div>
      </div>
      
      {!isJoined ? (
        <div className="rounded-lg bg-white p-8 shadow-sm">
          <div className="text-center">
            <h3 className="text-lg font-medium text-gray-900">Join this conversation</h3>
            <p className="mt-2 text-sm text-gray-500">
              You need to join this space to view messages and participate in the conversation.
            </p>
            <button
              onClick={handleJoin}
              className="mt-4 rounded-md bg-primary-600 px-4 py-2 text-white shadow-sm hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
            >
              Join Space
            </button>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-8 lg:grid-cols-4">
          <div className="lg:col-span-3">
            <div className="rounded-lg bg-white p-6 shadow-sm">
              <h2 className="mb-4 text-lg font-medium text-gray-900">Conversation</h2>
              
              <div className="mb-4 h-96 overflow-y-auto rounded-md bg-gray-50 p-4">
                {messagesLoading ? (
                  <div className="flex h-full items-center justify-center">
                    <p className="text-gray-500">Loading messages...</p>
                  </div>
                ) : localMessages.length > 0 ? (
                  <div className="space-y-4">
                    {localMessages.map((message: Message) => (
                      <div 
                        key={message.id} 
                        className={`rounded-lg p-3 shadow-sm ${
                          message.userId === user?.id ? 'ml-8 bg-primary-50' : 'mr-8 bg-white'
                        }`}
                      >
                        <div className="flex items-center gap-2">
                          <div 
                            className="h-8 w-8 rounded-full flex items-center justify-center text-white"
                            style={{ backgroundColor: message.userColor ? `var(--color-${message.userColor}-500)` : 'var(--color-primary-500)' }}
                          >
                            <span className="text-sm font-medium">
                              {message.userName.charAt(0).toUpperCase()}
                            </span>
                          </div>
                          <div>
                            <p className="text-sm font-medium text-gray-900">{message.userName}</p>
                            <p className="text-xs text-gray-500">
                              {formatDistanceToNow(new Date(message.createdAt), { addSuffix: true })}
                            </p>
                          </div>
                        </div>
                        <p className="mt-2 text-gray-700">{message.content}</p>
                      </div>
                    ))}
                    <div ref={messagesEndRef} />
                  </div>
                ) : (
                  <div className="flex h-full items-center justify-center">
                    <p className="text-gray-500">No messages yet. Start the conversation!</p>
                  </div>
                )}
              </div>
              
              {/* Typing indicators */}
              {Object.keys(typingUsers).length > 0 && (
                <div className="mb-2 text-xs text-gray-500 italic">
                  {Object.keys(typingUsers).length === 1 
                    ? `${typingUsers[Object.keys(typingUsers)[0]]} is typing...` 
                    : `${Object.keys(typingUsers).length} people are typing...`}
                </div>
              )}
              
              <form onSubmit={handleSendMessage}>
                <div className="flex gap-2">
                  <input
                    type="text"
                    name="message"
                    placeholder="Type your message..."
                    className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-primary-600 sm:text-sm sm:leading-6"
                    required
                    onChange={handleTyping}
                  />
                  <button
                    type="submit"
                    disabled={isSending}
                    className="rounded-md bg-primary-600 px-3.5 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
                  >
                    {isSending ? 'Sending...' : 'Send'}
                  </button>
                </div>
              </form>
            </div>
          </div>
          
          <div className="lg:col-span-1">
            <div className="rounded-lg bg-white p-6 shadow-sm">
              <h2 className="mb-4 text-lg font-medium text-gray-900">Space Info</h2>
              
              <div className="space-y-4">
                <div>
                  <h3 className="text-sm font-medium text-gray-500">Template Type</h3>
                  <p className="mt-1 text-sm text-gray-900 capitalize">{space.templateType.replace('_', ' ')}</p>
                </div>
                
                {space.isGeoLocal && space.location && (
                  <div>
                    <h3 className="text-sm font-medium text-gray-500">Location</h3>
                    <p className="mt-1 text-sm text-gray-900">
                      {space.locationRadius ? `Within ${space.locationRadius.toFixed(1)} km radius` : 'Local area'}
                    </p>
                  </div>
                )}
                
                {space.relatedSpaces && space.relatedSpaces.length > 0 && (
                  <div>
                    <h3 className="text-sm font-medium text-gray-500">Related Spaces</h3>
                    <ul className="mt-1 space-y-1">
                      {space.relatedSpaces.map((relatedId: string) => (
                        <li key={relatedId}>
                          <a href={`/spaces/${relatedId}`} className="text-sm text-primary-600 hover:text-primary-800">
                            Related Space #{relatedId.substring(0, 8)}
                          </a>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
                
                {space.expiresAt && (
                  <div>
                    <h3 className="text-sm font-medium text-gray-500">Expires</h3>
                    <p className="mt-1 text-sm text-gray-900">
                      {formatDistanceToNow(new Date(space.expiresAt), { addSuffix: true })}
                    </p>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
} 