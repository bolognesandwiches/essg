'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useSelector } from 'react-redux'
import { RootState } from '@/lib/store'
import { api } from '@/lib/api'
import { SpaceCard } from '@/components/SpaceCard'

export default function JoinedSpacesPage() {
  const router = useRouter()
  const { joinedSpaces } = useSelector((state: RootState) => state.spaces)
  
  const { data: spaces, isLoading, error } = api.useGetJoinedSpacesQuery()
  
  return (
    <div className="container-wide py-12">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Your Spaces</h1>
        <p className="mt-2 text-gray-600">
          Spaces you've joined and are participating in
        </p>
      </div>
      
      {isLoading ? (
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="h-64 animate-pulse rounded-lg bg-gray-200"></div>
          ))}
        </div>
      ) : error ? (
        <div className="rounded-md bg-red-50 p-4">
          <div className="flex">
            <div className="text-sm text-red-700">
              <p>Unable to load your spaces. Please try again later.</p>
            </div>
          </div>
        </div>
      ) : spaces && spaces.length > 0 ? (
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {spaces.map((space) => (
            <SpaceCard key={space.id} space={space} />
          ))}
        </div>
      ) : joinedSpaces.length > 0 ? (
        <div className="rounded-md bg-yellow-50 p-4">
          <div className="flex">
            <div className="text-sm text-yellow-700">
              <p>Loading your joined spaces...</p>
            </div>
          </div>
        </div>
      ) : (
        <div className="rounded-md bg-blue-50 p-4">
          <div className="flex">
            <div className="text-sm text-blue-700">
              <p>You haven't joined any spaces yet. Explore trending spaces to find conversations to join!</p>
            </div>
          </div>
        </div>
      )}
      
      <div className="mt-8 flex justify-center">
        <button
          onClick={() => router.push('/explore')}
          className="rounded-md bg-primary-600 px-4 py-2 text-white shadow-sm hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
        >
          Discover More Spaces
        </button>
      </div>
    </div>
  )
} 