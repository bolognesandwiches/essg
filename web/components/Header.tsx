'use client'

import { useState, useEffect, useRef } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useSelector, useDispatch } from 'react-redux'
import { RootState } from '@/lib/store'
import { updateDisplayName, resetUser } from '@/lib/features/user/userSlice'

export function Header() {
  const pathname = usePathname()
  const dispatch = useDispatch()
  const { user } = useSelector((state: RootState) => state.user)
  const { userLocation } = useSelector((state: RootState) => state.spaces)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [isEditingName, setIsEditingName] = useState(false)
  const [displayName, setDisplayName] = useState(user?.displayName || '')
  const inputRef = useRef<HTMLInputElement>(null)
  
  useEffect(() => {
    if (isEditingName && inputRef.current) {
      inputRef.current.focus()
    }
  }, [isEditingName])
  
  useEffect(() => {
    if (user) {
      setDisplayName(user.displayName)
    }
  }, [user])
  
  const handleSaveName = () => {
    if (displayName.trim()) {
      dispatch(updateDisplayName(displayName.trim()))
    }
    setIsEditingName(false)
  }
  
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSaveName()
    } else if (e.key === 'Escape') {
      setIsEditingName(false)
      setDisplayName(user?.displayName || '')
    }
  }
  
  const navigation = [
    { name: 'Home', href: '/' },
    { name: 'Explore', href: '/explore' },
    { name: 'Trends', href: '/trends' },
    { name: 'My Spaces', href: '/spaces/joined' },
  ]
  
  return (
    <header className="bg-white shadow">
      <nav className="mx-auto flex max-w-7xl items-center justify-between p-4 lg:px-8" aria-label="Global">
        <div className="flex lg:flex-1">
          <Link href="/" className="-m-1.5 p-1.5">
            <span className="sr-only">Ephemeral Social Space Generator</span>
            <div className="flex items-center">
              <div className="h-8 w-8 rounded-full bg-gradient-to-r from-primary-500 to-secondary-500"></div>
              <span className="ml-2 text-xl font-bold text-gray-900">ESSG</span>
            </div>
          </Link>
        </div>
        
        <div className="flex lg:hidden">
          <button
            type="button"
            className="-m-2.5 inline-flex items-center justify-center rounded-md p-2.5 text-gray-700"
            onClick={() => setMobileMenuOpen(true)}
          >
            <span className="sr-only">Open main menu</span>
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-6 w-6">
              <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
            </svg>
          </button>
        </div>
        
        <div className="hidden lg:flex lg:gap-x-12">
          {navigation.map((item) => (
            <Link
              key={item.name}
              href={item.href}
              className={`text-sm font-semibold ${
                pathname === item.href ? 'text-primary-600' : 'text-gray-900 hover:text-primary-600'
              }`}
            >
              {item.name}
            </Link>
          ))}
        </div>
        
        <div className="hidden lg:flex lg:flex-1 lg:justify-end">
          <div className="flex items-center">
            {userLocation && (
              <div className="mr-4 flex items-center text-sm text-gray-600">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="mr-1 h-4 w-4">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M15 10.5a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z" />
                </svg>
                Location enabled
              </div>
            )}
            
            {isEditingName ? (
              <div className="flex items-center">
                <input
                  ref={inputRef}
                  type="text"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  onKeyDown={handleKeyDown}
                  onBlur={handleSaveName}
                  className="mr-2 w-40 rounded-md border-gray-300 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
                  maxLength={20}
                />
                <button
                  onClick={handleSaveName}
                  className="rounded-md bg-primary-600 px-2 py-1 text-xs font-medium text-white hover:bg-primary-700"
                >
                  Save
                </button>
              </div>
            ) : (
              <div className="flex items-center">
                <button 
                  onClick={() => setIsEditingName(true)}
                  className="flex items-center"
                >
                  <div className="h-8 w-8 overflow-hidden rounded-full" style={{ backgroundColor: `var(--color-${user?.color}-500)` }}>
                    <div className="flex h-full w-full items-center justify-center text-white">
                      {user?.displayName.charAt(0).toUpperCase()}
                    </div>
                  </div>
                  <span className="ml-2 text-sm font-medium text-gray-900">{user?.displayName}</span>
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="ml-1 h-4 w-4 text-gray-500">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                  </svg>
                </button>
                <button
                  onClick={() => dispatch(resetUser())}
                  className="ml-4 text-xs text-gray-500 hover:text-gray-700"
                  title="Reset your anonymous identity"
                >
                  Reset
                </button>
              </div>
            )}
          </div>
        </div>
      </nav>
      
      {/* Mobile menu, show/hide based on mobile menu state */}
      {mobileMenuOpen && (
        <div className="lg:hidden">
          <div className="fixed inset-0 z-50" onClick={() => setMobileMenuOpen(false)}></div>
          <div className="fixed inset-y-0 right-0 z-50 w-full overflow-y-auto bg-white px-6 py-6 sm:max-w-sm sm:ring-1 sm:ring-gray-900/10">
            <div className="flex items-center justify-between">
              <Link href="/" className="-m-1.5 p-1.5" onClick={() => setMobileMenuOpen(false)}>
                <span className="sr-only">Ephemeral Social Space Generator</span>
                <div className="flex items-center">
                  <div className="h-8 w-8 rounded-full bg-gradient-to-r from-primary-500 to-secondary-500"></div>
                  <span className="ml-2 text-xl font-bold text-gray-900">ESSG</span>
                </div>
              </Link>
              <button
                type="button"
                className="-m-2.5 rounded-md p-2.5 text-gray-700"
                onClick={() => setMobileMenuOpen(false)}
              >
                <span className="sr-only">Close menu</span>
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-6 w-6">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <div className="mt-6 flow-root">
              <div className="-my-6 divide-y divide-gray-500/10">
                <div className="space-y-2 py-6">
                  {navigation.map((item) => (
                    <Link
                      key={item.name}
                      href={item.href}
                      className={`-mx-3 block rounded-lg px-3 py-2 text-base font-semibold leading-7 ${
                        pathname === item.href ? 'text-primary-600' : 'text-gray-900 hover:bg-gray-50'
                      }`}
                      onClick={() => setMobileMenuOpen(false)}
                    >
                      {item.name}
                    </Link>
                  ))}
                </div>
                <div className="py-6">
                  <div className="mb-4 flex items-center">
                    <div className="h-8 w-8 overflow-hidden rounded-full" style={{ backgroundColor: `var(--color-${user?.color}-500)` }}>
                      <div className="flex h-full w-full items-center justify-center text-white">
                        {user?.displayName.charAt(0).toUpperCase()}
                      </div>
                    </div>
                    <span className="ml-2 text-sm font-medium text-gray-900">{user?.displayName}</span>
                  </div>
                  
                  <div className="flex items-center">
                    <input
                      type="text"
                      value={displayName}
                      onChange={(e) => setDisplayName(e.target.value)}
                      placeholder="Change display name"
                      className="mr-2 w-full rounded-md border-gray-300 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
                      maxLength={20}
                    />
                    <button
                      onClick={() => {
                        if (displayName.trim()) {
                          dispatch(updateDisplayName(displayName.trim()))
                        }
                      }}
                      className="rounded-md bg-primary-600 px-2 py-1 text-xs font-medium text-white hover:bg-primary-700"
                    >
                      Save
                    </button>
                  </div>
                  
                  <button
                    onClick={() => {
                      dispatch(resetUser())
                      setMobileMenuOpen(false)
                    }}
                    className="mt-4 -mx-3 block w-full rounded-lg px-3 py-2.5 text-left text-base font-semibold leading-7 text-gray-900 hover:bg-gray-50"
                  >
                    Reset Identity
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </header>
  )
} 