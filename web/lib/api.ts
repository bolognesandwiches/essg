import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { BaseQueryFn, FetchArgs, FetchBaseQueryError } from '@reduxjs/toolkit/query'

// Define the base URL for API requests
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'

// Define types for API responses
export interface Space {
  id: string
  title: string
  description: string
  templateType: string
  lifecycleStage: string
  createdAt: string
  lastActive: string
  userCount: number
  messageCount: number
  isGeoLocal: boolean
  topicTags: string[]
  trendName?: string
  location?: {
    latitude: number
    longitude: number
  }
  locationRadius?: number
  relatedSpaces?: string[]
  expiresAt?: string
}

export interface Trend {
  id: string
  topic: string
  description: string
  keywords: string[]
  score: number
  velocity: number
  isGeoLocal: boolean
  firstDetected: string
  lastUpdated: string
  source: string
  url?: string
}

export interface SocialTrend {
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

export interface Message {
  id: string
  spaceId: string
  userId: string
  userName: string
  userColor?: string
  content: string
  createdAt: string
  replyToId?: string  // ID of the message this is replying to
  replyToUserName?: string // Username of the message this is replying to
}

// Create a custom base query with anonymous user ID
const baseQueryWithAnonymousId: BaseQueryFn<
  string | FetchArgs,
  unknown,
  FetchBaseQueryError
> = async (args, api, extraOptions) => {
  const baseQuery = fetchBaseQuery({ 
    baseUrl: API_BASE_URL,
    prepareHeaders: (headers) => {
      // Add anonymous user ID to headers if available
      if (typeof window !== 'undefined') {
        const anonymousUser = localStorage.getItem('anonymous_user')
        if (anonymousUser) {
          try {
            const user = JSON.parse(anonymousUser)
            headers.set('x-anonymous-user-id', user.id)
            headers.set('x-anonymous-user-name', user.displayName)
            headers.set('x-anonymous-user-color', user.color)
          } catch (e) {
            // Ignore parsing errors
          }
        }
      }
      return headers
    },
  })
  return baseQuery(args, api, extraOptions)
}

// Create the API service
export const api = createApi({
  reducerPath: 'api',
  baseQuery: baseQueryWithAnonymousId,
  tagTypes: ['Spaces', 'Trends', 'Messages', 'SocialTrends'],
  endpoints: (builder) => ({
    // Spaces endpoints
    getTrendingSpaces: builder.query<Space[], void>({
      query: () => 'spaces/trending',
      providesTags: ['Spaces'],
    }),
    getNearbySpaces: builder.query<Space[], { lat: number; lng: number; radius?: number }>({
      query: ({ lat, lng, radius = 10 }) => 
        `spaces/nearby?lat=${lat}&lng=${lng}&radius=${radius}`,
      providesTags: ['Spaces'],
    }),
    getSpaceById: builder.query<Space, string>({
      query: (id) => `spaces/${id}`,
      providesTags: (result, error, id) => [{ type: 'Spaces', id }],
    }),
    getJoinedSpaces: builder.query<Space[], void>({
      query: () => 'spaces/joined',
      providesTags: ['Spaces'],
    }),
    
    // Trends endpoints
    getTrendingTopics: builder.query<Trend[], void>({
      query: () => 'trends',
      providesTags: ['Trends'],
    }),
    
    // Social Media Trends endpoints
    getSocialTrends: builder.query<SocialTrend[], { source: string, subreddit?: string, location?: string, timeRange?: string, limit?: number }>({
      query: ({ source, subreddit, location, timeRange, limit }) => {
        // Build query parameters
        const params: Record<string, string> = { source };
        if (subreddit) params.subreddit = subreddit;
        if (location) params.location = location;
        if (timeRange) params.timeRange = timeRange;
        if (limit) params.limit = String(limit);
        
        return {
          url: `/social/trends`,
          params,
        };
      },
      providesTags: ['SocialTrends'],
      // Handle both regular responses and responses with mock data due to API limitations
      transformResponse: (response: any) => {
        console.log('Transforming social trends response:', response);
        
        // Check if response has error structure with mock data
        if (response && response.error && response.data) {
          console.warn(`Twitter API issue: ${response.error} - ${response.message}`);
          return response.data;
        }
        
        // Regular response
        return response;
      },
      transformErrorResponse: (response, meta, arg) => {
        console.error('Error fetching social trends:', response)
        return response
      },
    }),
    
    // Add the missing getAvailableLocations endpoint
    getAvailableLocations: builder.query<any[], { source: string, limit?: number }>({
      query: ({ source, limit }) => {
        // Build query parameters
        const params: Record<string, string> = { source };
        if (limit) params.limit = String(limit);
        
        return {
          url: `/social/locations`,
          params,
        };
      },
      providesTags: ['SocialTrends'],
    }),
    
    createSpaceFromTrend: builder.mutation<Space, { trendId: string, source: string }>({
      query: (body) => ({
        url: '/spaces',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Spaces', 'Trends'],
    }),
    
    // Add an endpoint to check if a space exists for a trend
    checkSpaceExistsForTrend: builder.query<{ exists: boolean, spaceId?: string }, { trendName: string, source: string }>({
      query: ({ trendName, source }) => 
        `/spaces/check-exists?trendName=${encodeURIComponent(trendName)}&source=${encodeURIComponent(source)}`,
    }),
    
    // Messages endpoints
    getSpaceMessages: builder.query<Message[], string>({
      query: (spaceId) => `spaces/${spaceId}/messages`,
      providesTags: (result, error, spaceId) => [{ type: 'Messages', id: spaceId }],
    }),
    sendMessage: builder.mutation<Message, { 
      spaceId: string; 
      content: string; 
      replyToId?: string;
      replyToUserName?: string;
    }>({
      query: ({ spaceId, content, replyToId, replyToUserName }) => ({
        url: `/spaces/${spaceId}/messages`,
        method: 'POST',
        body: { content, replyToId, replyToUserName },
      }),
      invalidatesTags: (result, error, { spaceId }) => [{ type: 'Messages', id: spaceId }],
    }),
  }),
})

// Export hooks for usage in components
export const {
  useGetSocialTrendsQuery,
  useGetAvailableLocationsQuery,
  useCreateSpaceFromTrendMutation,
  useCheckSpaceExistsForTrendQuery,
  useGetSpaceQuery,
  useGetSpacesQuery,
  useGetJoinedSpacesQuery,
  useGetNearbySpacesQuery,
  useJoinSpaceMutation,
  useLeaveSpaceMutation,
  useGetMessagesQuery,
  useSendMessageMutation,
} = api 