import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import type { Space } from '@/lib/api'

interface SpacesState {
  activeSpaceId: string | null
  joinedSpaces: string[]
  userLocation: {
    latitude: number
    longitude: number
  } | null
}

// Load joined spaces from localStorage
const loadJoinedSpaces = (): string[] => {
  if (typeof window !== 'undefined') {
    const storedSpaces = localStorage.getItem('joined_spaces')
    if (storedSpaces) {
      try {
        return JSON.parse(storedSpaces)
      } catch (e) {
        // If parsing fails, return empty array
        return []
      }
    }
  }
  return []
}

const initialState: SpacesState = {
  activeSpaceId: null,
  joinedSpaces: loadJoinedSpaces(),
  userLocation: null,
}

export const spacesSlice = createSlice({
  name: 'spaces',
  initialState,
  reducers: {
    setActiveSpace: (state, action: PayloadAction<string | null>) => {
      state.activeSpaceId = action.payload
    },
    joinSpace: (state, action: PayloadAction<string>) => {
      if (!state.joinedSpaces.includes(action.payload)) {
        state.joinedSpaces.push(action.payload)
        
        // Save to localStorage
        if (typeof window !== 'undefined') {
          localStorage.setItem('joined_spaces', JSON.stringify(state.joinedSpaces))
        }
      }
    },
    leaveSpace: (state, action: PayloadAction<string>) => {
      state.joinedSpaces = state.joinedSpaces.filter(id => id !== action.payload)
      if (state.activeSpaceId === action.payload) {
        state.activeSpaceId = null
      }
      
      // Update localStorage
      if (typeof window !== 'undefined') {
        localStorage.setItem('joined_spaces', JSON.stringify(state.joinedSpaces))
      }
    },
    setUserLocation: (state, action: PayloadAction<{ latitude: number; longitude: number } | null>) => {
      state.userLocation = action.payload
    },
  },
})

export const { 
  setActiveSpace, 
  joinSpace, 
  leaveSpace, 
  setUserLocation 
} = spacesSlice.actions

export default spacesSlice.reducer 