'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useSelector } from 'react-redux'
import { RootState } from '@/lib/store'
import { api } from '@/lib/api'

// Define types for social trends
interface SocialTrend {
  id: string
  name: string
  query: string
  tweet_volume: number
  source: string
  url?: string
}

// Define types for API query arguments
interface TrendsQueryArgs {
  source: string
  location?: string
}

export default function TrendsPage() {
  const router = useRouter()
  const { userLocation } = useSelector((state: RootState) => state.spaces)
  const [selectedSource, setSelectedSource] = useState('twitter')
  const [locationCode, setLocationCode] = useState<string | undefined>(undefined)
  const [debugInfo, setDebugInfo] = useState<string>('')
  const [isTestingConnection, setIsTestingConnection] = useState(false)
  
  // Fetch trends based on selected source and location
  const { 
    data: socialTrends, 
    isLoading: trendsLoading, 
    error: trendsError,
    refetch: refetchTrends
  } = api.useGetSocialTrendsQuery({
    source: selectedSource,
    location: locationCode
  }, {
    // Log the response for debugging
    onQueryStarted: async (arg: TrendsQueryArgs, { queryFulfilled }: { queryFulfilled: Promise<any> }) => {
      try {
        const { data } = await queryFulfilled
        console.log('Trends data received:', data)
        setDebugInfo('')
      } catch (error) {
        console.error('Error fetching trends:', error)
        setDebugInfo(JSON.stringify(error, null, 2))
      }
    }
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
  
  const handleRetry = () => {
    console.log('Retrying fetch...')
    setDebugInfo('Retrying...')
    refetchTrends()
  }
  
  // Function to test the debug endpoint
  const testDebugEndpoint = async () => {
    setIsTestingConnection(true)
    try {
      console.log('Testing Twitter API connection...')
      const response = await fetch('http://localhost:8080/api/social/debug')
      
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
          // Try to parse the body as JSON for better formatting
          const bodyJson = JSON.parse(data.body)
          formattedDebugInfo += `Body: ${JSON.stringify(bodyJson, null, 2)}`
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
      const response = await fetch('http://localhost:8080/api/social/trends?source=twitter')
      
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
      
      {/* Debug buttons */}
      <div className="mb-4 flex space-x-4">
        <button
          onClick={testDebugEndpoint}
          disabled={isTestingConnection}
          className="rounded-md bg-gray-200 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-300 disabled:opacity-50"
        >
          {isTestingConnection ? 'Testing...' : 'Test Twitter API Connection'}
        </button>
        <button
          onClick={testTrendsEndpoint}
          disabled={isTestingConnection}
          className="rounded-md bg-gray-200 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-300 disabled:opacity-50"
        >
          {isTestingConnection ? 'Testing...' : 'Test Trends Endpoint Directly'}
        </button>
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
          <div className="flex flex-col">
            <div className="text-sm text-red-700">
              <p>Unable to load trends. Please try again later.</p>
              <button 
                onClick={handleRetry} 
                className="mt-2 font-medium text-red-700 hover:text-red-600"
              >
                Retry
              </button>
            </div>
            {debugInfo && (
              <div className="mt-4">
                <p className="text-xs font-medium text-red-700">Debug Info:</p>
                <pre className="mt-1 max-h-40 overflow-auto rounded bg-red-100 p-2 text-xs text-red-800">
                  {debugInfo}
                </pre>
              </div>
            )}
          </div>
        </div>
      ) : !socialTrends || socialTrends.length === 0 ? (
        <div className="rounded-md bg-blue-50 p-4">
          <div className="flex flex-col">
            <div className="text-sm text-blue-700">
              <p>No trends available at the moment. Please try again later or select a different source.</p>
              <button 
                onClick={handleRetry} 
                className="mt-2 font-medium text-blue-700 hover:text-blue-600"
              >
                Retry
              </button>
            </div>
            {debugInfo && (
              <div className="mt-4">
                <p className="text-xs font-medium text-blue-700">Debug Info:</p>
                <pre className="mt-1 max-h-40 overflow-auto rounded bg-blue-100 p-2 text-xs text-blue-800">
                  {debugInfo}
                </pre>
              </div>
            )}
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          {socialTrends.map((trend: SocialTrend) => (
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
                <button
                  onClick={() => handleCreateSpace(trend)}
                  disabled={isCreating}
                  className="rounded-md bg-primary-500 px-4 py-2 text-sm font-medium text-white hover:bg-primary-600 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50"
                >
                  {isCreating ? 'Creating...' : 'Create Space'}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
} 