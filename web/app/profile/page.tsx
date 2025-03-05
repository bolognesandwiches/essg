'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useSelector, useDispatch } from 'react-redux'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { RootState } from '@/lib/store'
import { updateUserProfile } from '@/lib/features/user/userSlice'
import { api } from '@/lib/api'

// Form validation schema
const profileSchema = z.object({
  username: z.string().min(3, 'Username must be at least 3 characters'),
  email: z.string().email('Please enter a valid email address'),
})

type ProfileFormData = z.infer<typeof profileSchema>

export default function ProfilePage() {
  const router = useRouter()
  const dispatch = useDispatch()
  const { user, isAuthenticated } = useSelector((state: RootState) => state.user)
  const [isEditing, setIsEditing] = useState(false)
  const [updateSuccess, setUpdateSuccess] = useState(false)
  
  const { register, handleSubmit, formState: { errors }, reset } = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      username: user?.username || '',
      email: user?.email || '',
    }
  })
  
  // Redirect if not authenticated
  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login')
    }
  }, [isAuthenticated, router])
  
  // Update form when user data changes
  useEffect(() => {
    if (user) {
      reset({
        username: user.username,
        email: user.email,
      })
    }
  }, [user, reset])
  
  const onSubmit = async (data: ProfileFormData) => {
    try {
      // In a real app, you would call an API endpoint to update the user profile
      // For now, we'll just update the Redux state
      dispatch(updateUserProfile({
        username: data.username,
        email: data.email,
      }))
      
      setUpdateSuccess(true)
      setIsEditing(false)
      
      // Clear success message after 3 seconds
      setTimeout(() => {
        setUpdateSuccess(false)
      }, 3000)
    } catch (error) {
      console.error('Failed to update profile:', error)
    }
  }
  
  if (!user) {
    return (
      <div className="container-narrow py-12">
        <div className="animate-pulse">
          <div className="h-8 w-1/3 rounded bg-gray-200"></div>
          <div className="mt-4 h-4 w-2/3 rounded bg-gray-200"></div>
        </div>
      </div>
    )
  }
  
  return (
    <div className="container-narrow py-12">
      <div className="md:flex md:items-center md:justify-between">
        <div className="min-w-0 flex-1">
          <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl sm:tracking-tight">
            Your Profile
          </h2>
        </div>
        <div className="mt-4 flex md:ml-4 md:mt-0">
          <button
            type="button"
            onClick={() => setIsEditing(!isEditing)}
            className="ml-3 inline-flex items-center rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
          >
            {isEditing ? 'Cancel' : 'Edit Profile'}
          </button>
        </div>
      </div>
      
      {updateSuccess && (
        <div className="mt-4 rounded-md bg-green-50 p-4">
          <div className="flex">
            <div className="text-sm text-green-700">
              Profile updated successfully!
            </div>
          </div>
        </div>
      )}
      
      <div className="mt-8 overflow-hidden bg-white shadow sm:rounded-lg">
        {isEditing ? (
          <div className="px-4 py-5 sm:p-6">
            <form onSubmit={handleSubmit(onSubmit)}>
              <div className="space-y-6">
                <div>
                  <label htmlFor="username" className="block text-sm font-medium text-gray-700">
                    Username
                  </label>
                  <div className="mt-1">
                    <input
                      id="username"
                      type="text"
                      className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                      {...register('username')}
                    />
                    {errors.username && (
                      <p className="mt-1 text-sm text-red-600">{errors.username.message}</p>
                    )}
                  </div>
                </div>
                
                <div>
                  <label htmlFor="email" className="block text-sm font-medium text-gray-700">
                    Email address
                  </label>
                  <div className="mt-1">
                    <input
                      id="email"
                      type="email"
                      className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                      {...register('email')}
                    />
                    {errors.email && (
                      <p className="mt-1 text-sm text-red-600">{errors.email.message}</p>
                    )}
                  </div>
                </div>
                
                <div className="flex justify-end">
                  <button
                    type="button"
                    onClick={() => setIsEditing(false)}
                    className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="ml-3 inline-flex justify-center rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
                  >
                    Save
                  </button>
                </div>
              </div>
            </form>
          </div>
        ) : (
          <div className="border-t border-gray-200">
            <dl>
              <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="text-sm font-medium text-gray-500">Username</dt>
                <dd className="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">{user.username}</dd>
              </div>
              <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="text-sm font-medium text-gray-500">Email address</dt>
                <dd className="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">{user.email}</dd>
              </div>
              <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="text-sm font-medium text-gray-500">User ID</dt>
                <dd className="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">{user.id}</dd>
              </div>
            </dl>
          </div>
        )}
      </div>
      
      <div className="mt-8">
        <h3 className="text-lg font-medium leading-6 text-gray-900">Account Settings</h3>
        <div className="mt-4 overflow-hidden bg-white shadow sm:rounded-lg">
          <div className="border-t border-gray-200">
            <dl>
              <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="text-sm font-medium text-gray-500">Change Password</dt>
                <dd className="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">
                  <button
                    type="button"
                    className="text-primary-600 hover:text-primary-500"
                    onClick={() => router.push('/change-password')}
                  >
                    Update your password
                  </button>
                </dd>
              </div>
              <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="text-sm font-medium text-gray-500">Delete Account</dt>
                <dd className="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">
                  <button
                    type="button"
                    className="text-red-600 hover:text-red-500"
                    onClick={() => {
                      if (confirm('Are you sure you want to delete your account? This action cannot be undone.')) {
                        // In a real app, you would call an API endpoint to delete the account
                        console.log('Account deletion would happen here')
                      }
                    }}
                  >
                    Permanently delete your account
                  </button>
                </dd>
              </div>
            </dl>
          </div>
        </div>
      </div>
    </div>
  )
} 