'use client'

import { useState } from 'react'
import { useSelector, useDispatch } from 'react-redux'
import { RootState } from '@/lib/store'
import { api } from '@/lib/api'
import { SpaceCard } from '@/components/SpaceCard'
import { setUserLocation } from '@/lib/features/spaces/spacesSlice'
import { setLocationEnabled } from '@/lib/features/user/userSlice'

export default function ExplorePage() {
  const dispatch = useDispatch()
  const [activeTab, setActiveTab] = useState<'trending' | 'nearby'>('trending')
  const { locationEnabled } = useSelector((state: RootState) => state.user)
  const { userLocation } = useSelector((state: RootState) => state.spaces)
  const [searchRadius, setSearchRadius] = useState(10) // Default 10km radius
  const [locationLoading, setLocationLoading] = useState(false)
  
  const { data: trendingSpaces, isLoading: trendingLoading } = api.useGetTrendingSpacesQuery()
  
  const { data: nearbySpaces, isLoading: nearbyLoading } = api.useGetNearbySpacesQuery(
    userLocation && locationEnabled 
      ? { 
          lat: userLocation.latitude, 
          lng: userLocation.longitude,
          radius: searchRadius
        } 
      : { lat: 0, lng: 0 }, // Dummy values when location not available
    { 
      skip: !locationEnabled || !userLocation || activeTab !== 'nearby'
    }
  )
  
  const enableLocation = () => {
    setLocationLoading(true)
    
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (position) => {
          dispatch(setUserLocation({
            latitude: position.coords.latitude,
            longitude: position.coords.longitude
          }))
          dispatch(setLocationEnabled(true))
          setLocationLoading(false)
        },
        (error) => {
          console.error('Error getting location:', error)
          setLocationLoading(false)
        }
      )
    } else {
      console.error('Geolocation is not supported by this browser')
      setLocationLoading(false)
    }
  }
  
  return (
    <div className="container-wide py-12">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Explore Spaces</h1>
        <p className="mt-2 text-gray-600">
          Discover trending conversations and join temporary spaces that evolve naturally
        </p>
      </div>
      
      <div className="mb-8 border-b border-gray-200">
        <nav className="-mb-px flex space-x-8" aria-label="Tabs">
          <button
            onClick={() => setActiveTab('trending')}
            className={`${
              activeTab === 'trending'
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
            } whitespace-nowrap border-b-2 px-1 py-4 text-sm font-medium`}
          >
            Trending
          </button>
          <button
            onClick={() => setActiveTab('nearby')}
            className={`${
              activeTab === 'nearby'
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
            } whitespace-nowrap border-b-2 px-1 py-4 text-sm font-medium`}
          >
            Nearby
          </button>
        </nav>
      </div>
      
      {activeTab === 'nearby' && (
        <div className="mb-6">
          {!locationEnabled || !userLocation ? (
            <div className="rounded-md bg-yellow-50 p-4">
              <div className="flex">
                <div className="text-sm text-yellow-700">
                  <p>Enable location to see nearby conversations.</p>
                </div>
              </div>
            </div>
          ) : (
            <div className="flex items-center space-x-4">
              <label htmlFor="radius" className="text-sm font-medium text-gray-700">
                Search radius: {searchRadius} km
              </label>
              <input
                type="range"
                id="radius"
                name="radius"
                min="1"
                max="50"
                value={searchRadius}
                onChange={(e) => setSearchRadius(parseInt(e.target.value))}
                className="h-2 w-full max-w-md rounded-lg appearance-none bg-gray-200"
              />
            </div>
          )}
        </div>
      )}
      
      {activeTab === 'trending' ? (
        <>
          {trendingLoading ? (
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {[...Array(6)].map((_, i) => (
                <div key={i} className="h-64 animate-pulse rounded-lg bg-gray-200"></div>
              ))}
            </div>
          ) : !trendingSpaces || trendingSpaces.length === 0 ? (
            <div className="rounded-md bg-blue-50 p-4">
              <div className="flex">
                <div className="text-sm text-blue-700">
                  <p>No trending spaces found at the moment. Check back soon!</p>
                </div>
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {trendingSpaces.map((space) => (
                <SpaceCard key={space.id} space={space} />
              ))}
            </div>
          )}
        </>
      ) : (
        <>
          {!locationEnabled || !userLocation ? (
            <div className="flex h-64 items-center justify-center rounded-lg border-2 border-dashed border-gray-300 bg-white p-12 text-center">
              <div>
                <svg
                  className="mx-auto h-12 w-12 text-gray-400"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  aria-hidden="true"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
                  />
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
                  />
                </svg>
                <h3 className="mt-2 text-sm font-medium text-gray-900">Location access required</h3>
                <p className="mt-1 text-sm text-gray-500">
                  Enable location access to discover nearby conversations.
                </p>
                <div className="mt-6">
                  <button
                    type="button"
                    onClick={enableLocation}
                    disabled={locationLoading}
                    className="inline-flex items-center rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
                  >
                    {locationLoading ? 'Detecting location...' : 'Enable location'}
                  </button>
                </div>
              </div>
            </div>
          ) : nearbyLoading ? (
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {[...Array(6)].map((_, i) => (
                <div key={i} className="h-64 animate-pulse rounded-lg bg-gray-200"></div>
              ))}
            </div>
          ) : !nearbySpaces || nearbySpaces.length === 0 ? (
            <div className="rounded-md bg-blue-50 p-4">
              <div className="flex">
                <div className="text-sm text-blue-700">
                  <p>No nearby spaces found within {searchRadius}km. Try increasing the search radius or check back later!</p>
                </div>
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {nearbySpaces.map((space) => (
                <SpaceCard key={space.id} space={space} />
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
} 