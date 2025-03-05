'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useDispatch, useSelector } from 'react-redux'
import { setUserLocation } from '@/lib/features/spaces/spacesSlice'
import { setLocationEnabled } from '@/lib/features/user/userSlice'
import { RootState } from '@/lib/store'

export function Hero() {
  const dispatch = useDispatch()
  const { locationEnabled } = useSelector((state: RootState) => state.user)
  const { userLocation } = useSelector((state: RootState) => state.spaces)
  const [locationLoading, setLocationLoading] = useState(false)
  
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
    <div className="bg-gradient-to-r from-primary-600 to-secondary-600 px-6 py-24 text-center sm:py-32">
      <div className="mx-auto max-w-2xl">
        <h1 className="text-4xl font-bold tracking-tight text-white sm:text-6xl">
          Join the conversation where it matters
        </h1>
        <p className="mt-6 text-lg leading-8 text-gray-100">
          Discover trending discussions and join temporary spaces that evolve naturally.
          No endless feeds, no accounts required - just meaningful conversations that come and go like they do in real life.
        </p>
        <div className="mt-10 flex items-center justify-center gap-x-6">
          <Link 
            href="/explore" 
            className="rounded-md bg-white px-3.5 py-2.5 text-sm font-semibold text-primary-600 shadow-sm hover:bg-gray-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white"
          >
            Explore Spaces
          </Link>
          
          {locationEnabled && userLocation ? (
            <div className="flex items-center text-sm font-semibold text-white">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="mr-1 h-5 w-5">
                <path strokeLinecap="round" strokeLinejoin="round" d="M15 10.5a3 3 0 11-6 0 3 3 0 016 0z" />
                <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z" />
              </svg>
              Location enabled
            </div>
          ) : (
            <button
              onClick={enableLocation}
              disabled={locationLoading}
              className="text-sm font-semibold leading-6 text-white hover:text-gray-200"
            >
              {locationLoading ? 'Detecting location...' : 'Enable location'} <span aria-hidden="true">→</span>
            </button>
          )}
        </div>
      </div>
    </div>
  )
} 