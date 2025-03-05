'use client'

import { useSelector } from 'react-redux'
import { api } from '@/lib/api'
import { SpaceCard } from './SpaceCard'
import { RootState } from '@/lib/store'

export function NearbySpaces() {
  const { locationEnabled, userLocation } = useSelector((state: RootState) => state.spaces)
  
  const { data: spaces, isLoading, error } = api.useGetNearbySpacesQuery(
    userLocation && locationEnabled 
      ? { 
          lat: userLocation.latitude, 
          lng: userLocation.longitude,
          radius: 10 // 10km radius by default
        } 
      : { lat: 0, lng: 0 }, // Dummy values when location not available
    { 
      skip: !locationEnabled || !userLocation 
    }
  )
  
  if (!locationEnabled || !userLocation) {
    return (
      <div className="rounded-md bg-yellow-50 p-4">
        <div className="flex">
          <div className="text-sm text-yellow-700">
            <p>Enable location to see nearby conversations.</p>
          </div>
        </div>
      </div>
    )
  }
  
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {[...Array(3)].map((_, i) => (
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
            <p>Unable to load nearby spaces. Please try again later.</p>
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
            <p>No nearby spaces found at the moment. Be the first to start a local conversation!</p>
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
    </div>
  )
} 