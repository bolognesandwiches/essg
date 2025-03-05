import Link from 'next/link'
import { TrendingSpaces } from '@/components/TrendingSpaces'
import { NearbySpaces } from '@/components/NearbySpaces'
import { Hero } from '@/components/Hero'

export default function Home() {
  return (
    <main>
      <Hero />
      
      <section className="container-wide py-12">
        <div className="mb-8">
          <h2 className="text-2xl font-bold text-gray-900">Trending Conversations</h2>
          <p className="mt-2 text-gray-600">
            Join these active discussion spaces based on trending topics
          </p>
        </div>
        
        <TrendingSpaces />
        
        <div className="mt-12 mb-8">
          <h2 className="text-2xl font-bold text-gray-900">Nearby Conversations</h2>
          <p className="mt-2 text-gray-600">
            Discover what people are talking about in your area
          </p>
        </div>
        
        <NearbySpaces />
      </section>
    </main>
  )
} 