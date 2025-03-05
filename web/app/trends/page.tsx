'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useSelector } from 'react-redux'
import { RootState } from '@/lib/store'
import { api, SocialTrend } from '@/lib/api'

export default function TrendsPage() {
  const router = useRouter()
  const { userLocation } = useSelector((state: RootState) => state.spaces)
  const [selectedSource, setSelectedSource] = useState('twitter')
  const [locationCode, setLocationCode] = useState<string | undefined>(undefined)
  
  // Fetch trends based on selected source and location
  const { 
    data: socialTrends, 
    isLoading: trendsLoading, 
    error: trendsError,
    refetch: refetchTrends
  } = api.useGetSocialTrendsQuery({
    source: selectedSource,
    location: locationCode
  })
  
  // Create space from trend
  const [createSpace, { isLoading: isCreating }] = api.useCreateSpaceFromTrendMutation()
  
  // Update location code when user location changes
  useEffect(() => {
    if (userLocation) {
      // This is a simplified example - in a real implementation,
      // you might want to reverse geocode the coordinates to get a location code
      // that the Twitter API understands (like a WOEID)
      setLocationCode('1') // Default to worldwide if we can't determine location
    }
  }, [userLocation])
  
  const handleCreateSpace = async (trend: SocialTrend) => {
    try {
      const space = await createSpace({ 
        trendId: trend.id, 
        source: trend.source 
      }).unwrap()
      
      // Navigate to the newly created space
      router.push(`/spaces/${space.id}`)
    } catch (error) {
      console.error('Failed to create space:', error)
    }
  }
  
  const sources = [
    { id: 'twitter', name: 'Twitter/X' },
    // Add more sources as they become available
    { id: 'reddit', name: 'Reddit', disabled: true },
    { id: 'mastodon', name: 'Mastodon', disabled: true },
  ]
  
  return (
    <div className="container-wide py-12">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Trending Topics</h1>
        <p className="mt-2 text-gray-600">
          Discover trending topics from social media and create ephemeral spaces around them
        </p>
      </div>
      
      {/* Source selector */}
      <div className="mb-8">
        <label htmlFor="source-selector" className="block text-sm font-medium text-gray-700">
          Source
        </label>
        <div className="mt-1">
          <select
            id="source-selector"
            value={selectedSource}
            onChange={(e) => setSelectedSource(e.target.value)}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
          >
            {sources.map((source) => (
              <option 
                key={source.id} 
                value={source.id}
                disabled={source.disabled}
              >
                {source.name} {source.disabled ? '(Coming Soon)' : ''}
              </option>
            ))}
          </select>
        </div>
      </div>
      
      {/* Trends list */}
      {trendsLoading ? (
        <div className="space-y-4">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="animate-pulse rounded-lg bg-white p-6 shadow-sm">
              <div className="h-6 w-1/3 rounded bg-gray-200"></div>
              <div className="mt-2 h-4 w-2/3 rounded bg-gray-200"></div>
              <div className="mt-4 h-10 w-1/4 rounded bg-gray-200"></div>
            </div>
          ))}
        </div>
      ) : trendsError ? (
        <div className="rounded-md bg-red-50 p-4">
          <div className="flex">
            <div className="text-sm text-red-700">
              <p>Unable to load trends. Please try again later.</p>
              <button 
                onClick={() => refetchTrends()} 
                className="mt-2 font-medium text-red-700 hover:text-red-600"
              >
                Retry
              </button>
            </div>
          </div>
        </div>
      ) : !socialTrends || socialTrends.length === 0 ? (
        <div className="rounded-md bg-blue-50 p-4">
          <div className="flex">
            <div className="text-sm text-blue-700">
              <p>No trends available at the moment. Please try again later or select a different source.</p>
            </div>
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          {socialTrends.map((trend) => (
            <div key={trend.id} className="rounded-lg bg-white p-6 shadow-sm">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium text-gray-900">{trend.name}</h3>
                  {trend.tweet_volume && (
                    <p className="mt-1 text-sm text-gray-500">
                      {trend.tweet_volume.toLocaleString()} tweets
                    </p>
                  )}
                </div>
                <div className="flex items-center space-x-4">
                  {trend.url && (
                    <a
                      href={trend.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm font-medium text-gray-500 hover:text-gray-700"
                    >
                      View on {trend.source === 'twitter' ? 'Twitter/X' : trend.source}
                    </a>
                  )}
                  <button
                    onClick={() => handleCreateSpace(trend)}
                    disabled={isCreating}
                    className="rounded-md bg-primary-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
                  >
                    {isCreating ? 'Creating...' : 'Create Space'}
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
} 