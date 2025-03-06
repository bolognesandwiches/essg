'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useSelector } from 'react-redux'
import { RootState } from '@/lib/store'
import { api } from '@/lib/api'
import Image from 'next/image'
import { formatDistanceToNow } from 'date-fns'
import { InformationCircleIcon } from '@heroicons/react/24/outline'

// Define types for social trends
interface SocialTrend {
  id: string
  name: string
  query: string
  score: number
  comments_count: number
  source: string
  url?: string
  subreddit?: string
  thumbnail?: string
  author?: string
  created?: number
}

// Define a type for subreddit data
interface Subreddit {
  display_name: string
  title: string
  subscribers: number
  public_description?: string
  url?: string
  icon_img?: string
}

// Define types for API query arguments
interface TrendsQueryArgs {
  source: string
  subreddit?: string
  location?: string
  timeRange?: string
  limit?: number
}

export default function TrendsPage() {
  const router = useRouter()
  const { userLocation } = useSelector((state: RootState) => state.spaces)
  const [selectedSource, setSelectedSource] = useState('reddit')
  const [selectedSubreddit, setSelectedSubreddit] = useState('')
  const [selectedLocation, setSelectedLocation] = useState('')
  const [timeRange, setTimeRange] = useState('day')
  const [debugInfo, setDebugInfo] = useState<string>('')
  const [isTestingConnection, setIsTestingConnection] = useState(false)
  const [apiErrorMessage, setApiErrorMessage] = useState<string>('')
  
  // State to track which trends already have spaces
  const [existingSpaces, setExistingSpaces] = useState<Record<string, string>>({})
  
  // Fetch trends based on selected source and location
  const { 
    data: socialTrends, 
    isLoading: trendsLoading, 
    error: trendsError,
    refetch: refetchTrends
  } = api.useGetSocialTrendsQuery({
    source: selectedSource,
    subreddit: selectedSource === 'reddit' ? selectedSubreddit : undefined,
    location: selectedSource === 'twitter' ? selectedLocation : undefined,
    timeRange: selectedSource === 'reddit' ? timeRange : undefined
  }, {
    // Log the response for debugging
    onQueryStarted: async (arg: TrendsQueryArgs, { queryFulfilled }: { queryFulfilled: Promise<any> }) => {
      try {
        const { data } = await queryFulfilled
        console.log('Trends data received:', data)
        setDebugInfo('')
        setApiErrorMessage('')
      } catch (error: any) {
        console.error('Error fetching trends:', error)
        
        // Extract and format the error message for display
        let errorMessage = 'Error fetching trends'
        if (error?.error?.data) {
          errorMessage = error.error.data
        } else if (error?.message) {
          errorMessage = error.message
        }
        
        if (selectedSource === 'twitter' && errorMessage.includes('403')) {
          errorMessage = 'Twitter API access is limited. Using mock trends data instead.'
        }
        
        setApiErrorMessage(errorMessage)
        setDebugInfo(JSON.stringify(error, null, 2))
      }
    }
  })
  
  // Create space from trend
  const [createSpace, { isLoading: isCreating, error: createSpaceError }] = api.useCreateSpaceFromTrendMutation()
  
  // Fetch locations based on selected source
  const { 
    data: locations, 
    isLoading: locationsLoading 
  } = api.useGetAvailableLocationsQuery({
    source: selectedSource
  })
  
  // Effect to reset specific selectors when source changes
  useEffect(() => {
    if (selectedSource === 'reddit') {
      setSelectedLocation('')
    } else if (selectedSource === 'twitter') {
      setSelectedSubreddit('')
    }
  }, [selectedSource])
  
  // Function to check if a space exists for a trend
  const checkSpaceExists = async (trend: SocialTrend) => {
    try {
      const result = await fetch(`http://localhost:8080/api/spaces/check-exists?trendName=${encodeURIComponent(trend.name)}&source=${encodeURIComponent(trend.source)}`)
      if (result.ok) {
        const data = await result.json()
        if (data.exists && data.spaceId) {
          setExistingSpaces(prev => ({
            ...prev,
            [trend.id]: data.spaceId
          }))
        }
      }
    } catch (error) {
      console.error('Error checking if space exists:', error)
    }
  }
  
  // Check if spaces exist for all trends when they load
  useEffect(() => {
    if (socialTrends && socialTrends.length > 0) {
      // Reset existing spaces
      setExistingSpaces({})
      
      // Check each trend
      socialTrends.forEach((trend: SocialTrend) => {
        checkSpaceExists(trend)
      })
    }
  }, [socialTrends])
  
  const handleCreateOrJoinSpace = async (trend: SocialTrend) => {
    try {
      console.log('Creating/joining space for trend:', trend)
      setDebugInfo(`Creating/joining space from ${trend.source} trend: ${trend.name}`)
      
      // Check if space already exists for this trend
      const existingSpaceId = existingSpaces[trend.id]
      
      if (existingSpaceId) {
        // Space exists, navigate to it
        console.log('Space already exists, navigating to it:', existingSpaceId)
        router.push(`/spaces/${existingSpaceId}`)
        return
      }
      
      // Create a new space
      const trendId = trend.name
      
      const space = await createSpace({ 
        trendId: trendId, // Use the trend name/title instead of id
        source: trend.source 
      }).unwrap()
      
      console.log('Space created:', space)
      setDebugInfo('')
      
      // Navigate to the newly created space
      router.push(`/spaces/${space.id}`)
    } catch (error) {
      console.error('Failed to create/join space:', error)
      setDebugInfo(`Error creating/joining space: ${JSON.stringify(error, null, 2)}`)
    }
  }
  
  const sources = [
    { id: 'reddit', name: 'Reddit' },
    { id: 'twitter', name: 'Twitter/X' },
    // Add more sources as they become available
    { id: 'mastodon', name: 'Mastodon', disabled: true },
  ]
  
  const timeRanges = [
    { id: 'hour', name: 'Past Hour' },
    { id: 'day', name: 'Today' },
    { id: 'week', name: 'This Week' },
    { id: 'month', name: 'This Month' },
    { id: 'year', name: 'This Year' },
    { id: 'all', name: 'All Time' },
  ]
  
  const handleRetry = () => {
    console.log('Retrying fetch...')
    setDebugInfo('Retrying...')
    refetchTrends()
  }
  
  // Function to test the debug endpoint
  const testDebugEndpoint = async () => {
    setIsTestingConnection(true)
    try {
      console.log('Testing Reddit API connection...')
      const response = await fetch('http://localhost:8080/api/social/debug?source=reddit')
      
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`)
      }
      
      const data = await response.json()
      console.log('Debug endpoint response:', data)
      
      // Format the debug info nicely
      let formattedDebugInfo = `Status Code: ${data.status_code}\n\n`
      
      // Add response body if it exists
      if (data.body) {
        try {
          // If body is already an object, stringify it
          if (typeof data.body === 'object') {
            formattedDebugInfo += `Body: ${JSON.stringify(data.body, null, 2)}`
          } else {
            // Try to parse the body as JSON for better formatting
            const bodyJson = JSON.parse(data.body)
            formattedDebugInfo += `Body: ${JSON.stringify(bodyJson, null, 2)}`
          }
        } catch (e) {
          // If parsing fails, just use the string
          formattedDebugInfo += `Body: ${data.body}`
        }
      }
      
      setDebugInfo(formattedDebugInfo)
    } catch (error) {
      console.error('Error testing debug endpoint:', error)
      setDebugInfo(`Error: ${error instanceof Error ? error.message : String(error)}`)
    } finally {
      setIsTestingConnection(false)
    }
  }
  
  // Function to directly test the trends endpoint
  const testTrendsEndpoint = async () => {
    setIsTestingConnection(true)
    try {
      console.log('Testing trends endpoint directly...')
      const response = await fetch(`http://localhost:8080/api/social/trends?source=reddit&subreddit=${selectedSubreddit}&timeRange=${timeRange}`)
      
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`)
      }
      
      const data = await response.json()
      console.log('Trends endpoint response:', data)
      setDebugInfo(JSON.stringify(data, null, 2))
    } catch (error) {
      console.error('Error testing trends endpoint:', error)
      setDebugInfo(`Error: ${error instanceof Error ? error.message : String(error)}`)
    } finally {
      setIsTestingConnection(false)
    }
  }
  
  // Format the date from a Unix timestamp
  const formatDate = (timestamp?: number) => {
    if (!timestamp) return '';
    const date = new Date(timestamp * 1000);
    return date.toLocaleString();
  }
  
  // Effect to clear error message when source changes
  useEffect(() => {
    setApiErrorMessage('');
  }, [selectedSource]);
  
  return (
    <div className="container-wide py-12">
      <div className="mb-8 flex flex-col md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Explore Trending Topics</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
            Discover and join conversations about trending topics across the web
          </p>
        </div>
        
        {/* Source selection */}
        <div className="mt-4 md:mt-0">
          <div className="flex flex-wrap gap-2">
            <button
              onClick={() => setSelectedSource('reddit')}
              className={`rounded-full px-4 py-2 text-sm font-medium ${
                selectedSource === 'reddit' 
                  ? 'bg-primary-100 text-primary-800 dark:bg-primary-900/30 dark:text-primary-300' 
                  : 'bg-gray-100 text-gray-800 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700'
              }`}
            >
              Reddit
            </button>
            <button
              onClick={() => setSelectedSource('twitter')}
              className={`rounded-full px-4 py-2 text-sm font-medium ${
                selectedSource === 'twitter' 
                  ? 'bg-primary-100 text-primary-800 dark:bg-primary-900/30 dark:text-primary-300' 
                  : 'bg-gray-100 text-gray-800 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700'
              }`}
            >
              Twitter
            </button>
          </div>
        </div>
      </div>
      
      {/* API Error message display */}
      {apiErrorMessage && (
        <div className="mb-6 rounded-md bg-amber-50 dark:bg-amber-900/20 p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <InformationCircleIcon className="h-5 w-5 text-amber-500 dark:text-amber-400" aria-hidden="true" />
            </div>
            <div className="ml-3">
              <p className="text-sm text-amber-700 dark:text-amber-300">{apiErrorMessage}</p>
              {selectedSource === 'twitter' && apiErrorMessage.includes('Twitter API') && (
                <p className="mt-1 text-xs text-amber-600 dark:text-amber-400">
                  Note: Twitter API access now requires a paid developer account. The application is showing mock data.
                </p>
              )}
            </div>
          </div>
        </div>
      )}
      
      {/* Source selector */}
      <div className="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div>
          <label htmlFor="source-selector" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
            Source
          </label>
          <div className="mt-1">
            <select
              id="source-selector"
              value={selectedSource}
              onChange={(e) => setSelectedSource(e.target.value)}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
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
        
        {/* Reddit specific controls */}
        {selectedSource === 'reddit' && (
          <>
            <div>
              <label htmlFor="subreddit-selector" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Subreddit
              </label>
              <div className="mt-1">
                <select
                  id="subreddit-selector"
                  value={selectedSubreddit}
                  onChange={(e) => setSelectedSubreddit(e.target.value)}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
                  disabled={locationsLoading}
                >
                  <option value="">r/popular (Default)</option>
                  {locations && Array.isArray(locations) && locations.map((subreddit: any) => (
                    <option key={subreddit.display_name} value={subreddit.display_name}>
                      r/{subreddit.display_name} ({subreddit.subscribers?.toLocaleString() || 'N/A'} members)
                    </option>
                  ))}
                </select>
              </div>
            </div>
            
            <div>
              <label htmlFor="time-selector" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Time Range
              </label>
              <div className="mt-1">
                <select
                  id="time-selector"
                  value={timeRange}
                  onChange={(e) => setTimeRange(e.target.value)}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
                >
                  {timeRanges.map((range) => (
                    <option key={range.id} value={range.id}>
                      {range.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </>
        )}
        
        {/* Twitter specific controls */}
        {selectedSource === 'twitter' && (
          <div>
            <label htmlFor="location-selector" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Location
            </label>
            <div className="mt-1">
              <select
                id="location-selector"
                value={selectedLocation}
                onChange={(e) => setSelectedLocation(e.target.value)}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
                disabled={locationsLoading}
              >
                <option value="">Worldwide (Default)</option>
                {locations && Array.isArray(locations) && locations.map((location: any) => (
                  <option key={location.woeid} value={location.woeid}>
                    {location.name}, {location.countryCode}
                  </option>
                ))}
              </select>
            </div>
          </div>
        )}
      </div>
      
      {/* Debug and test buttons */}
      <div className="mb-4 flex space-x-4">
        <button
          onClick={handleRetry}
          className="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200 dark:hover:bg-gray-600"
          disabled={trendsLoading}
        >
          ↻ Refresh Trends
        </button>
        
        <button
          onClick={testTrendsEndpoint}
          className="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200 dark:hover:bg-gray-600"
          disabled={isTestingConnection}
        >
          Test API Connection
        </button>
      </div>
      
      {/* Trends list */}
      {trendsLoading ? (
        <div className="space-y-4">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="animate-pulse rounded-lg bg-white p-6 shadow-sm dark:bg-gray-800">
              <div className="h-6 w-1/3 rounded bg-gray-200 dark:bg-gray-700"></div>
              <div className="mt-2 h-4 w-2/3 rounded bg-gray-200 dark:bg-gray-700"></div>
              <div className="mt-4 h-10 w-1/4 rounded bg-gray-200 dark:bg-gray-700"></div>
            </div>
          ))}
        </div>
      ) : trendsError ? (
        <div className="rounded-md bg-red-50 p-4 dark:bg-red-900/20">
          <div className="flex flex-col">
            <div className="text-sm text-red-700 dark:text-red-400">
              <p>Unable to load trends. Please try again later.</p>
              <button 
                onClick={handleRetry} 
                className="mt-2 font-medium text-red-700 hover:text-red-600 dark:text-red-400 dark:hover:text-red-300"
              >
                Retry
              </button>
            </div>
            {debugInfo && (
              <div className="mt-4">
                <p className="text-xs font-medium text-red-700 dark:text-red-400">Debug Info:</p>
                <pre className="mt-1 max-h-40 overflow-auto rounded bg-red-100 p-2 text-xs text-red-800 dark:bg-red-900/40 dark:text-red-300">
                  {debugInfo}
                </pre>
              </div>
            )}
          </div>
        </div>
      ) : !socialTrends || socialTrends.length === 0 ? (
        <div className="rounded-md bg-blue-50 p-4 dark:bg-blue-900/20">
          <div className="flex flex-col">
            <div className="text-sm text-blue-700 dark:text-blue-400">
              <p>No trends available at the moment. Please try again later or select a different source.</p>
              <button 
                onClick={handleRetry} 
                className="mt-2 font-medium text-blue-700 hover:text-blue-600 dark:text-blue-400 dark:hover:text-blue-300"
              >
                Retry
              </button>
            </div>
            {debugInfo && (
              <div className="mt-4">
                <p className="text-xs font-medium text-blue-700 dark:text-blue-400">Debug Info:</p>
                <pre className="mt-1 max-h-40 overflow-auto rounded bg-blue-100 p-2 text-xs text-blue-800 dark:bg-blue-900/40 dark:text-blue-300">
                  {debugInfo}
                </pre>
              </div>
            )}
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          {socialTrends.map((trend: SocialTrend) => (
            <div key={trend.id} className="rounded-lg bg-white p-6 shadow-sm dark:bg-gray-800">
              <div className="flex items-start">
                {trend.source === 'reddit' && trend.thumbnail && trend.thumbnail !== 'self' && trend.thumbnail !== 'default' && (
                  <div className="mr-4 flex-shrink-0">
                    <img 
                      src={trend.thumbnail} 
                      alt="" 
                      className="h-20 w-20 rounded object-cover"
                      onError={(e) => {
                        // Hide the image on error
                        (e.target as HTMLImageElement).style.display = 'none'
                      }}
                    />
                  </div>
                )}
                <div className="flex-1">
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="text-lg font-medium text-gray-900 dark:text-white">{trend.name}</h3>
                      {trend.source === 'reddit' && (
                        <div className="mt-1 flex items-center text-sm text-gray-500 dark:text-gray-400">
                          <span className="mr-3">r/{trend.subreddit}</span>
                          <span className="mr-3">👤 {trend.author}</span>
                          <span>📅 {formatDate(trend.created)}</span>
                        </div>
                      )}
                      <div className="mt-2 flex items-center text-sm text-gray-500 dark:text-gray-400">
                        {trend.source === 'reddit' ? (
                          <>
                            <span className="mr-3">👍 {trend.score.toLocaleString()}</span>
                            <span>💬 {trend.comments_count.toLocaleString()}</span>
                          </>
                        ) : (
                          <span>{trend.score.toLocaleString()} {trend.source === 'twitter' ? 'tweets' : 'points'}</span>
                        )}
                      </div>
                    </div>
                    <button
                      onClick={() => handleCreateOrJoinSpace(trend)}
                      disabled={isCreating}
                      className="rounded-md bg-primary-500 px-4 py-2 text-sm font-medium text-white hover:bg-primary-600 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:bg-primary-600 dark:hover:bg-primary-700 dark:focus:ring-offset-gray-800 disabled:opacity-50"
                    >
                      {isCreating ? 'Processing...' : existingSpaces[trend.id] ? 'Join Space' : 'Create Space'}
                    </button>
                  </div>
                  
                  {trend.url && (
                    <div className="mt-2">
                      <a 
                        href={trend.url} 
                        target="_blank" 
                        rel="noopener noreferrer"
                        className="text-sm text-blue-500 hover:underline dark:text-blue-400"
                      >
                        View original post
                      </a>
                    </div>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
} 