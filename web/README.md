# ESSG Frontend

This is the frontend application for the Ephemeral Social Space Generator (ESSG), built with Next.js and deployed on Vercel.

## Getting Started

### Prerequisites

- Node.js 16+ and npm

### Installation

1. Install dependencies:
   ```bash
   npm install
   ```

2. Create a `.env.local` file in the root directory with the following variables:
   ```
   NEXT_PUBLIC_API_URL=http://localhost:8080/api
   ```

3. Start the development server:
   ```bash
   npm run dev
   ```

4. Open [http://localhost:3000](http://localhost:3000) in your browser to see the application.

## Project Structure

```
web/
├── app/                  # Next.js App Router
│   ├── layout.tsx        # Root layout component
│   ├── page.tsx          # Home page
│   ├── spaces/           # Space-related pages
│   │   └── [id]/         # Dynamic space detail page
│   └── globals.css       # Global styles
├── components/           # Reusable UI components
│   ├── Header.tsx        # Site header with navigation
│   ├── Hero.tsx          # Hero section for landing page
│   ├── SpaceCard.tsx     # Card component for displaying spaces
│   ├── TrendingSpaces.tsx # Trending spaces component
│   └── NearbySpaces.tsx  # Nearby spaces component
├── lib/                  # Application logic
│   ├── api.ts            # API client using RTK Query
│   ├── store.ts          # Redux store configuration
│   └── features/         # Redux slices and features
│       ├── spaces/       # Space-related state management
│       └── user/         # User-related state management
├── public/               # Static assets
└── styles/               # Additional styles
```

## Key Features

- **Real-time Space Discovery**: Browse trending and nearby conversation spaces
- **Location-based Filtering**: Find discussions relevant to your location
- **Interactive Messaging**: Join spaces and participate in conversations
- **Responsive Design**: Works on desktop and mobile devices

## Technologies Used

- **Next.js**: React framework for server-rendered applications
- **TypeScript**: Type-safe JavaScript
- **Redux Toolkit**: State management with RTK Query for API calls
- **Tailwind CSS**: Utility-first CSS framework
- **Socket.io-client**: Real-time WebSocket communication
- **Mapbox GL**: Interactive maps for location-based features

## Deployment

The application is configured for deployment on Vercel:

```bash
npm run build
vercel
```

## Development Guidelines

- Use TypeScript for all new components and functions
- Follow the existing component structure and naming conventions
- Use Tailwind CSS for styling
- Implement responsive designs that work on mobile and desktop
- Use RTK Query for all API calls
- Keep components small and focused on a single responsibility

## Backend Integration

The frontend communicates with the Go backend API. Make sure the backend is running and accessible at the URL specified in your `.env.local` file.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 