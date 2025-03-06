'use client'

import { useEffect, useState, useRef } from 'react'
import { useParams } from 'next/navigation'
import { useDispatch, useSelector } from 'react-redux'
import { formatDistanceToNow, parseISO } from 'date-fns'
import { api, Message } from '@/lib/api'
import { setActiveSpace, joinSpace } from '@/lib/features/spaces/spacesSlice'
import { RootState } from '@/lib/store'
import { useWebSocket, TypingIndicator } from '@/lib/websocket'

// Helper function to safely format dates
const safeFormatDate = (dateString: string | undefined) => {
  if (!dateString) return 'N/A';
  
  try {
    // Check if the input is a string or already a Date object
    const date = typeof dateString === 'string' 
      ? parseISO(dateString) 
      : dateString;
    
    // Check if the date is valid
    if (date instanceof Date && isNaN(date.getTime())) {
      console.error('Invalid date:', dateString);
      return 'N/A';
    }
    
    // Special handling for dates that are too far in the past or future
    const now = new Date();
    const dateObj = date instanceof Date ? date : new Date(date);
    
    // Check if date is more than 100 years in the past or future (likely an error)
    if (Math.abs(now.getFullYear() - dateObj.getFullYear()) > 100) {
      console.error('Date appears to be too far in the past/future:', dateString);
      return 'just now'; // Default to something reasonable
    }
    
    return formatDistanceToNow(dateObj, { addSuffix: true });
  } catch (error) {
    console.error('Error parsing date:', error, dateString);
    return 'N/A';
  }
}

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
  
  const [messageInput, setMessageInput] = useState('')
  const [replyTo, setReplyTo] = useState<{ id: string, userName: string, content: string } | null>(null)
  
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
  
  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!user?.id || !spaceId) return
    
    const form = e.target as HTMLFormElement
    const formData = new FormData(form)
    const message = formData.get('message') as string
    
    if (!message.trim()) return
    
    try {
      await sendMessage({
        spaceId,
        content: message,
        replyToId: replyTo?.id,
        replyToUserName: replyTo?.userName
      }).unwrap()
      
      // Clear the form and reply state
      form.reset()
      setMessageInput('')
      setReplyTo(null)
      
      // Stop typing indicator
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current)
        typingTimeoutRef.current = null
      }
      setIsTyping(false)
      webSocket.sendTypingIndicator(spaceId)
      
      // Scroll to bottom
      setTimeout(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
      }, 100)
    } catch (error) {
      console.error('Failed to send message:', error)
    }
  }
  
  const handleReply = (message: Message) => {
    // Truncate the message content if it's too long
    const truncatedContent = message.content.length > 50 
      ? message.content.substring(0, 50) + '...' 
      : message.content;
      
    setReplyTo({
      id: message.id,
      userName: message.userName,
      content: truncatedContent
    })
    // Focus the input
    const inputEl = document.querySelector('input[name="message"]') as HTMLInputElement
    if (inputEl) {
      inputEl.focus()
    }
  }
  
  const cancelReply = () => {
    setReplyTo(null)
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
        
        {space.trendName && space.trendName !== space.title && (
          <p className="mt-2 text-sm text-primary-600 font-medium">
            Trending topic: {space.trendName}
          </p>
        )}
        
        <div className="mt-4 flex flex-wrap gap-2">
          {space.topicTags && space.topicTags.length > 0 ? (
            space.topicTags.map((tag: string) => (
              <span key={tag} className="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-800">
                {tag}
              </span>
            ))
          ) : (
            <span className="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-800">
              General Discussion
            </span>
          )}
        </div>
        
        <div className="mt-4 flex items-center gap-4 text-sm text-gray-500">
          <span>{space.userCount || 0} participants</span>
          <span>{space.messageCount || 0} messages</span>
          <span>Created {safeFormatDate(space.createdAt)}</span>
          <span>Last active {safeFormatDate(space.lastActive)}</span>
          {space.expiresAt && (
            <span>Expires {safeFormatDate(space.expiresAt)}</span>
          )}
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
                        className={`relative rounded-lg p-4 shadow-sm mb-4 ${
                          message.userId === user?.id ? 'ml-8 bg-primary-50' : 'mr-8 bg-white'
                        }`}
                      >
                        {message.replyToUserName && (
                          <div className="text-xs text-gray-500 mb-1 flex items-center">
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-3 w-3 mr-1 rotate-180" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                            </svg>
                            Replying to <span className="font-medium">{message.replyToUserName}</span>
                          </div>
                        )}
                        
                        <div className="flex items-center gap-2">
                          <div 
                            className="h-8 w-8 rounded-full flex items-center justify-center text-white"
                            style={{ backgroundColor: message.userColor ? `var(--color-${message.userColor}-500)` : 'var(--color-primary-500)' }}
                          >
                            <span className="text-sm font-medium">
                              {(message.userName || 'User').charAt(0).toUpperCase()}
                            </span>
                          </div>
                          <div>
                            <p className="text-sm font-medium text-gray-900">{message.userName || 'Anonymous User'}</p>
                            <p className="text-xs text-gray-500">
                              {safeFormatDate(message.createdAt)}
                            </p>
                          </div>
                          
                          <button
                            onClick={() => handleReply(message)}
                            className="ml-auto text-xs text-gray-500 hover:text-primary-600"
                            title="Reply to this message"
                          >
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                            </svg>
                          </button>
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
                {replyTo && (
                  <div className="mb-2 flex items-center bg-gray-50 p-2 rounded-md">
                    <div className="flex flex-col text-xs">
                      <span className="text-gray-600">
                        Replying to <span className="font-medium">{replyTo.userName}</span>
                      </span>
                      <span className="text-gray-500 italic mt-1">{replyTo.content}</span>
                    </div>
                    <button 
                      type="button"
                      onClick={cancelReply}
                      className="ml-auto text-xs text-gray-500 hover:text-red-600"
                      title="Cancel reply"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                )}
                <div className="flex gap-2">
                  <input
                    type="text"
                    name="message"
                    placeholder="Type your message..."
                    className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-primary-600 sm:text-sm sm:leading-6"
                    required
                    value={messageInput}
                    onChange={(e) => {
                      setMessageInput(e.target.value)
                      handleTyping(e)
                    }}
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
                  <p className="mt-1 text-sm text-gray-900 capitalize">{(space.templateType || 'standard').replace('_', ' ')}</p>
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
                      {safeFormatDate(space.expiresAt)}
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