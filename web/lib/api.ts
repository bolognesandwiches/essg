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
  tweet_volume: number
  source: string
  url?: string
}

export interface Message {
  id: string
  spaceId: string
  userId: string
  userName: string
  userColor?: string
  content: string
  createdAt: string
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
    getSocialTrends: builder.query<SocialTrend[], { source: string, location?: string }>({
      query: ({ source, location }) => ({
        url: `/social/trends`,
        params: { source, location },
      }),
      providesTags: ['SocialTrends'],
      transformErrorResponse: (response, meta, arg) => {
        console.error('Error fetching social trends:', response)
        return response
      },
    }),
    createSpaceFromTrend: builder.mutation<Space, { trendId: string, source: string }>({
      query: (body) => ({
        url: '/spaces',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Spaces', 'Trends'],
    }),
    
    // Messages endpoints
    getSpaceMessages: builder.query<Message[], string>({
      query: (spaceId) => `spaces/${spaceId}/messages`,
      providesTags: (result, error, spaceId) => [{ type: 'Messages', id: spaceId }],
    }),
    sendMessage: builder.mutation<Message, { spaceId: string; content: string }>({
      query: ({ spaceId, content }) => ({
        url: `/spaces/${spaceId}/messages`,
        method: 'POST',
        body: { content },
      }),
      invalidatesTags: (result, error, { spaceId }) => [{ type: 'Messages', id: spaceId }],
    }),
  }),
})

// Export hooks for usage in components
export const {
  useGetSocialTrendsQuery,
  useCreateSpaceFromTrendMutation,
  useGetSpaceByIdQuery,
  useGetJoinedSpacesQuery,
  useGetTrendingSpacesQuery,
  useGetNearbySpacesQuery,
  useGetSpaceMessagesQuery,
  useSendMessageMutation,
  useLeaveSpaceMutation,
} = api 