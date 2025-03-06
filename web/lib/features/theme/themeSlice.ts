import { createSlice, PayloadAction } from '@reduxjs/toolkit'

interface ThemeState {
  darkMode: boolean
}

// Function to load theme preference from localStorage
const loadThemePreference = (): boolean => {
  if (typeof window !== 'undefined') {
    // Check localStorage first
    const storedPreference = localStorage.getItem('theme')
    if (storedPreference) {
      return storedPreference === 'dark'
    }
    
    // Check system preference
    return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
  }
  return false
}

// Initial state
const initialState: ThemeState = {
  darkMode: loadThemePreference()
}

export const themeSlice = createSlice({
  name: 'theme',
  initialState,
  reducers: {
    toggleDarkMode: (state) => {
      state.darkMode = !state.darkMode
      
      // Save to localStorage
      if (typeof window !== 'undefined') {
        localStorage.setItem('theme', state.darkMode ? 'dark' : 'light')
      }
    },
    setDarkMode: (state, action: PayloadAction<boolean>) => {
      state.darkMode = action.payload
      
      // Save to localStorage
      if (typeof window !== 'undefined') {
        localStorage.setItem('theme', state.darkMode ? 'dark' : 'light')
      }
    }
  }
})

export const { toggleDarkMode, setDarkMode } = themeSlice.actions

export default themeSlice.reducer 