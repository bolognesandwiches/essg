'use client'

import Link from 'next/link'
import { useDispatch } from 'react-redux'
import { formatDistanceToNow } from 'date-fns'
import { Space } from '@/lib/api'
import { joinSpace } from '@/lib/features/spaces/spacesSlice'

// Map template types to colors and icons
const templateConfig: Record<string, { bgColor: string, icon: React.ReactNode }> = {
  general: {
    bgColor: 'bg-blue-100',
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-blue-600">
        <path strokeLinecap="round" strokeLinejoin="round" d="M8.625 12a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    )
  },
  breaking_news: {
    bgColor: 'bg-red-100',
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-600">
        <path strokeLinecap="round" strokeLinejoin="round" d="M12 7.5h1.5m-1.5 3h1.5m-7.5 3h7.5m-7.5 3h7.5m3-9h3.375c.621 0 1.125.504 1.125 1.125V18a2.25 2.25 0 01-2.25 2.25M16.5 7.5V18a2.25 2.25 0 002.25 2.25M16.5 7.5V4.875c0-.621-.504-1.125-1.125-1.125H4.125C3.504 3.75 3 4.254 3 4.875V18a2.25 2.25 0 002.25 2.25h13.5M6 7.5h3v3H6v-3z" />
      </svg>
    )
  },
  event: {
    bgColor: 'bg-green-100',
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-green-600">
        <path strokeLinecap="round" strokeLinejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" />
      </svg>
    )
  },
  discussion: {
    bgColor: 'bg-purple-100',
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-purple-600">
        <path strokeLinecap="round" strokeLinejoin="round" d="M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 01-.825-.242m9.345-8.334a2.126 2.126 0 00-.476-.095 48.64 48.64 0 00-8.048 0c-1.131.094-1.976 1.057-1.976 2.192v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0011.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155" />
      </svg>
    )
  },
  local: {
    bgColor: 'bg-yellow-100',
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-yellow-600">
        <path strokeLinecap="round" strokeLinejoin="round" d="M15 10.5a3 3 0 11-6 0 3 3 0 016 0z" />
        <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z" />
      </svg>
    )
  }
}

// Get configuration for a template type with fallback to general
const getTemplateConfig = (templateType: string) => {
  return templateConfig[templateType] || templateConfig.general
}

// Map lifecycle stages to visual indicators
const getLifecycleIndicator = (stage: string) => {
  switch (stage) {
    case 'growing':
      return { color: 'text-green-500', label: 'Growing' }
    case 'peak':
      return { color: 'text-blue-500', label: 'Active' }
    case 'waning':
      return { color: 'text-yellow-500', label: 'Slowing' }
    case 'dissolving':
      return { color: 'text-red-500', label: 'Ending Soon' }
    default:
      return { color: 'text-gray-500', label: 'New' }
  }
}

interface SpaceCardProps {
  space: Space
}

export function SpaceCard({ space }: SpaceCardProps) {
  const dispatch = useDispatch()
  const templateConfig = getTemplateConfig(space.templateType)
  const lifecycle = getLifecycleIndicator(space.lifecycleStage)
  
  const handleJoin = () => {
    dispatch(joinSpace(space.id))
  }
  
  return (
    <div className="flex h-64 flex-col overflow-hidden rounded-lg bg-white shadow transition hover:shadow-md">
      <div className={`flex items-center gap-2 p-4 ${templateConfig.bgColor}`}>
        {templateConfig.icon}
        <span className="text-sm font-medium capitalize">
          {space.templateType.replace('_', ' ')}
        </span>
        <span className="ml-auto text-xs font-medium">
          <span className={`inline-block h-2 w-2 rounded-full ${lifecycle.color}`}></span>
          <span className="ml-1">{lifecycle.label}</span>
        </span>
      </div>
      
      <div className="flex flex-1 flex-col p-4">
        <h3 className="text-lg font-semibold text-gray-900">{space.title}</h3>
        <p className="mt-1 line-clamp-2 flex-1 text-sm text-gray-500">{space.description}</p>
        
        <div className="mt-4 flex flex-wrap gap-1">
          {space.topicTags.slice(0, 3).map((tag) => (
            <span key={tag} className="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-800">
              {tag}
            </span>
          ))}
          {space.topicTags.length > 3 && (
            <span className="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-800">
              +{space.topicTags.length - 3} more
            </span>
          )}
        </div>
        
        <div className="mt-4 flex items-center justify-between text-xs text-gray-500">
          <div className="flex items-center gap-2">
            <span>{space.userCount} participants</span>
            <span>•</span>
            <span>{space.messageCount} messages</span>
          </div>
          <span>
            {formatDistanceToNow(new Date(space.lastActive), { addSuffix: true })}
          </span>
        </div>
        
        <div className="mt-4 flex gap-2">
          <Link 
            href={`/spaces/${space.id}`}
            className="flex-1 rounded-md bg-white px-2.5 py-1.5 text-center text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
          >
            Preview
          </Link>
          <button
            onClick={handleJoin}
            className="flex-1 rounded-md bg-primary-600 px-2.5 py-1.5 text-center text-sm font-semibold text-white shadow-sm hover:bg-primary-700"
          >
            Join
          </button>
        </div>
      </div>
    </div>
  )
} 