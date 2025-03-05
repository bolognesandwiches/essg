import { configureStore } from '@reduxjs/toolkit'
import { setupListeners } from '@reduxjs/toolkit/query'
import { api } from './api'
import spacesReducer from './features/spaces/spacesSlice'
import userReducer from './features/user/userSlice'

export const store = configureStore({
  reducer: {
    [api.reducerPath]: api.reducer,
    spaces: spacesReducer,
    user: userReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(api.middleware),
})

setupListeners(store.dispatch)

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch 