# Ephemeral Social Space Generator (ESSG)

A platform for anonymous, location-based conversations that come and go naturally.

## Project Structure

- `server/` - Go backend
- `web/` - Next.js frontend

## Setup Instructions

### Prerequisites

- Go 1.18 or higher
- Node.js 16 or higher
- npm or yarn

### Backend Setup

1. Navigate to the server directory:
   ```
   cd server
   ```

2. Copy the example environment file:
   ```
   cp .env.example .env
   ```

3. Edit the `.env` file and add your Twitter Bearer Token:
   ```
   TWITTER_BEARER_TOKEN=your_twitter_bearer_token_here
   ```

4. Install Go dependencies:
   ```
   go mod tidy
   ```

5. Run the server:
   ```
   go run main.go
   ```

### Frontend Setup

1. Navigate to the web directory:
   ```
   cd web
   ```

2. Install dependencies:
   ```
   npm install
   ```
   or
   ```
   yarn
   ```

3. Create a `.env.local` file:
   ```
   NEXT_PUBLIC_API_URL=http://localhost:8080/api
   NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
   ```

4. Run the development server:
   ```
   npm run dev
   ```
   or
   ```
   yarn dev
   ```

5. Open [http://localhost:3000](http://localhost:3000) in your browser.

## Features

- Anonymous user system with random IDs and display names
- Location-based space discovery
- Real-time messaging with WebSockets
- Integration with social media trends
- Ephemeral spaces that evolve naturally

## Environment Variables

### Backend (.env)

- `TWITTER_BEARER_TOKEN` - Twitter API Bearer Token for fetching trends
- `PORT` - Port for the server to listen on (default: 8080)
- `ENV` - Environment (development, production)

### Frontend (.env.local)

- `NEXT_PUBLIC_API_URL` - URL for the backend API
- `NEXT_PUBLIC_WS_URL` - URL for the WebSocket connection