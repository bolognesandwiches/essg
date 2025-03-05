'use client'

import { useEffect } from 'react'
import Link from 'next/link'
import { useDispatch } from 'react-redux'
import { api } from '@/lib/api'
import { SpaceCard } from './SpaceCard'

export function TrendingSpaces() {
  const { data: spaces, isLoading, error } = api.useGetTrendingSpacesQuery()
  
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {[...Array(6)].map((_, i) => (
          <div key={i} className="h-64 animate-pulse rounded-lg bg-gray-200"></div>
        ))}
      </div>
    )
  }
  
  if (error) {
    return (
      <div className="rounded-md bg-red-50 p-4">
        <div className="flex">
          <div className="text-sm text-red-700">
            <p>Unable to load trending spaces. Please try again later.</p>
          </div>
        </div>
      </div>
    )
  }
  
  if (!spaces || spaces.length === 0) {
    return (
      <div className="rounded-md bg-blue-50 p-4">
        <div className="flex">
          <div className="text-sm text-blue-700">
            <p>No trending spaces found at the moment. Check back soon!</p>
          </div>
        </div>
      </div>
    )
  }
  
  return (
    <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
      {spaces.map((space) => (
        <SpaceCard key={space.id} space={space} />
      ))}
      
      <Link 
        href="/explore" 
        className="flex h-64 items-center justify-center rounded-lg border-2 border-dashed border-gray-300 bg-white p-6 text-center hover:border-gray-400 hover:bg-gray-50"
      >
        <div>
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-gray-100">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-6 w-6 text-gray-600">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
            </svg>
          </div>
          <h3 className="mt-2 text-sm font-semibold text-gray-900">Explore more spaces</h3>
          <p className="mt-1 text-sm text-gray-500">Discover all trending conversations</p>
        </div>
      </Link>
    </div>
  )
} 