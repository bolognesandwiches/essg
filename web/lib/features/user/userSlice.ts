import { createSlice, PayloadAction } from '@reduxjs/toolkit'

// Anonymous user with optional display name
export interface AnonymousUser {
  id: string
  displayName: string
  color: string
}

interface UserState {
  user: AnonymousUser | null
  locationEnabled: boolean
}

// Generate a random user ID and color for anonymous users
const generateAnonymousUser = (): AnonymousUser => {
  const randomId = Math.random().toString(36).substring(2, 15)
  const colors = [
    'red', 'blue', 'green', 'purple', 'orange', 
    'pink', 'teal', 'indigo', 'yellow', 'cyan'
  ]
  const randomColor = colors[Math.floor(Math.random() * colors.length)]
  
  return {
    id: randomId,
    displayName: `User_${randomId.substring(0, 5)}`,
    color: randomColor
  }
}

// Create anonymous user on initialization
const createInitialUser = (): AnonymousUser => {
  if (typeof window !== 'undefined') {
    // Check if we already have a user in localStorage
    const storedUser = localStorage.getItem('anonymous_user')
    if (storedUser) {
      try {
        return JSON.parse(storedUser)
      } catch (e) {
        // If parsing fails, create a new user
      }
    }
    
    // Create new anonymous user
    const newUser = generateAnonymousUser()
    localStorage.setItem('anonymous_user', JSON.stringify(newUser))
    return newUser
  }
  
  // Fallback for server-side rendering
  return generateAnonymousUser()
}

const initialState: UserState = {
  user: createInitialUser(),
  locationEnabled: false
}

export const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    updateDisplayName: (state, action: PayloadAction<string>) => {
      if (state.user) {
        state.user.displayName = action.payload
        
        // Update in localStorage
        if (typeof window !== 'undefined') {
          localStorage.setItem('anonymous_user', JSON.stringify(state.user))
        }
      }
    },
    resetUser: (state) => {
      const newUser = generateAnonymousUser()
      state.user = newUser
      
      // Update in localStorage
      if (typeof window !== 'undefined') {
        localStorage.setItem('anonymous_user', JSON.stringify(newUser))
      }
    },
    setLocationEnabled: (state, action: PayloadAction<boolean>) => {
      state.locationEnabled = action.payload
    }
  },
})

export const { 
  updateDisplayName,
  resetUser,
  setLocationEnabled
} = userSlice.actions

export default userSlice.reducer 