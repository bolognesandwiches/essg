'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useDispatch, useSelector } from 'react-redux'
import { RootState } from '@/lib/store'
import { api, Space } from '@/lib/api'
import { SpaceCard } from '@/components/SpaceCard'
import { joinSpace } from '@/lib/features/spaces/spacesSlice'

export default function JoinedSpacesPage() {
  const router = useRouter()
  const dispatch = useDispatch()
  const { joinedSpaces: joinedSpaceIds } = useSelector((state: RootState) => state.spaces)
  const [allSpaces, setAllSpaces] = useState<Space[]>([])
  
  const { 
    data: serverJoinedSpaces, 
    isLoading: serverSpacesLoading, 
    error: serverSpacesError,
    refetch: refetchServerSpaces
  } = api.useGetJoinedSpacesQuery()
  
  // Fetch individual spaces for each ID in Redux store that aren't in server response
  const fetchJoinedSpacesFromIds = async (serverSpaces: Space[] = []) => {
    if (joinedSpaceIds.length === 0) return []
    
    // Create a set of server space IDs for quick lookup
    const serverSpaceIds = new Set(serverSpaces.map(space => space.id))
    
    // Only fetch spaces that are in the joined list but not in server response
    const idsToFetch = joinedSpaceIds.filter((id: string) => !serverSpaceIds.has(id))
    
    if (idsToFetch.length === 0) return []
    
    try {
      // Create an array of promises for each space fetch
      const spacePromises = idsToFetch.map((id: string) => 
        fetch(`http://localhost:8080/api/spaces/${id}`)
          .then(res => res.ok ? res.json() : null)
          .catch(err => {
            console.error(`Error fetching space ${id}:`, err)
            return null
          })
      )
      
      // Wait for all promises to resolve
      const spaces = await Promise.all(spacePromises)
      return spaces.filter(Boolean) // Filter out any null responses
    } catch (error) {
      console.error('Error fetching joined spaces from IDs:', error)
      return []
    }
  }
  
  useEffect(() => {
    // Combine server spaces with client-side joined spaces
    const loadAllJoinedSpaces = async () => {
      // Start with server spaces if available
      const serverSpaces = serverJoinedSpaces || []
      
      // Add spaces from Redux store IDs that aren't already in the server response
      const additionalSpaces = await fetchJoinedSpacesFromIds(serverSpaces)
      
      // Combine spaces
      const combinedSpaces = [...serverSpaces]
      
      if (additionalSpaces && additionalSpaces.length > 0) {
        const existingIds = new Set(combinedSpaces.map(space => space.id))
        // Only add spaces that are in the joinedSpaceIds list
        const newSpaces = additionalSpaces.filter((space: Space) => 
          !existingIds.has(space.id) && joinedSpaceIds.includes(space.id)
        )
        combinedSpaces.push(...newSpaces)
      }
      
      // Update local state with combined spaces
      setAllSpaces(combinedSpaces)
      
      // Make sure Redux store has all of these spaces (don't add new ones)
      // Only sync IDs that are already in the joined list
      serverSpaces.forEach((space: Space) => {
        if (joinedSpaceIds.includes(space.id)) {
          dispatch(joinSpace(space.id))
        }
      })
    }
    
    loadAllJoinedSpaces()
  }, [serverJoinedSpaces, joinedSpaceIds, dispatch])
  
  const handleRetry = () => {
    refetchServerSpaces()
  }
  
  const isLoading = serverSpacesLoading && allSpaces.length === 0
  const error = serverSpacesError && allSpaces.length === 0
  
  return (
    <div className="container-wide py-12">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Your Spaces</h1>
        <p className="mt-2 text-gray-600 dark:text-gray-400">
          Spaces you've joined and are participating in
        </p>
      </div>
      
      {isLoading ? (
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="h-80 animate-pulse rounded-lg bg-gray-200 dark:bg-gray-700"></div>
          ))}
        </div>
      ) : error ? (
        <div className="rounded-md bg-red-50 dark:bg-red-900/20 p-4">
          <div className="flex flex-col">
            <div className="text-sm text-red-700 dark:text-red-400">
              <p>Unable to load your spaces. Please try again later.</p>
              <button 
                onClick={handleRetry} 
                className="mt-2 font-medium text-red-700 dark:text-red-400 hover:text-red-600 dark:hover:text-red-300"
              >
                Retry
              </button>
            </div>
          </div>
        </div>
      ) : allSpaces.length > 0 ? (
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 auto-rows-fr">
          {allSpaces.map((space: Space) => (
            <SpaceCard key={space.id} space={space} />
          ))}
        </div>
      ) : (
        <div className="rounded-md bg-blue-50 dark:bg-blue-900/20 p-4">
          <div className="flex">
            <div className="text-sm text-blue-700 dark:text-blue-400">
              <p>You haven't joined any spaces yet. Explore trending spaces to find conversations to join!</p>
            </div>
          </div>
        </div>
      )}
      
      <div className="mt-8 flex justify-center">
        <button
          onClick={() => router.push('/trends')}
          className="rounded-md bg-primary-600 px-4 py-2 text-white shadow-sm hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:bg-primary-700 dark:hover:bg-primary-600 dark:focus:ring-offset-gray-800"
        >
          Discover Trending Topics
        </button>
      </div>
    </div>
  )
} 