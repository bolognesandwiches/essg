📦essg
 ┣ 📂cmd
 ┃ ┣ 📂admin
 ┃ ┣ 📂api
 ┃ ┃ ┗ 📜main.go
 ┃ ┗ 📂worker
 ┣ 📂docs
 ┃ ┣ 📂api
 ┃ ┣ 📂architecture
 ┃ ┣ 📂deployment
 ┃ ┗ 📂development
 ┣ 📂internal
 ┃ ┣ 📂adapter
 ┃ ┃ ┣ 📂geo
 ┃ ┃ ┣ 📂queue
 ┃ ┃ ┣ 📂social
 ┃ ┃ ┗ 📂storage
 ┃ ┃ ┃ ┣ 📜space_store.go
 ┃ ┃ ┃ ┗ 📜trend_store.go
 ┃ ┣ 📂config
 ┃ ┃ ┗ 📜config.go
 ┃ ┣ 📂domain
 ┃ ┃ ┣ 📂geo
 ┃ ┃ ┃ ┗ 📜service.go
 ┃ ┃ ┣ 📂identity
 ┃ ┃ ┃ ┗ 📜service.go
 ┃ ┃ ┣ 📂messaging
 ┃ ┃ ┃ ┗ 📜service.go
 ┃ ┃ ┣ 📂space
 ┃ ┃ ┃ ┣ 📜manager.go
 ┃ ┃ ┃ ┗ 📜model.go
 ┃ ┃ ┗ 📂trend
 ┃ ┃ ┃ ┣ 📜detector.go
 ┃ ┃ ┃ ┗ 📜model.go
 ┃ ┣ 📂server
 ┃ ┃ ┣ 📂handlers
 ┃ ┃ ┃ ┣ 📜geo.go
 ┃ ┃ ┃ ┣ 📜space.go
 ┃ ┃ ┃ ┣ 📜trend.go
 ┃ ┃ ┃ ┗ 📜websocket.go
 ┃ ┃ ┗ 📜server.go
 ┃ ┗ 📂service
 ┃ ┃ ┣ 📂engagement
 ┃ ┃ ┣ 📂geo
 ┃ ┃ ┃ ┗ 📜service.go
 ┃ ┃ ┣ 📂identity
 ┃ ┃ ┣ 📂listening
 ┃ ┃ ┃ ┣ 📜analyzer.go
 ┃ ┃ ┃ ┣ 📜detector.go
 ┃ ┃ ┃ ┗ 📜geotagger.go
 ┃ ┃ ┣ 📂messaging
 ┃ ┃ ┗ 📂space
 ┃ ┃ ┃ ┣ 📜engagement.go
 ┃ ┃ ┃ ┣ 📜manager.go
 ┃ ┃ ┃ ┗ 📜templates.go
 ┣ 📂pkg
 ┃ ┣ 📂ephemeralid
 ┃ ┣ 📂geoutils
 ┃ ┣ 📂lifecycle
 ┃ ┗ 📂trendanalysis
 ┣ 📂scripts
 ┃ ┗ 📜schema.sql
 ┣ 📂test
 ┃ ┣ 📂e2e
 ┃ ┗ 📂integration
 ┣ 📂web
 ┃ ┣ 📂components
 ┃ ┣ 📂pages
 ┃ ┣ 📂public
 ┃ ┗ 📂styles
 ┣ 📜docker-compose.yml
 ┣ 📜Dockerfile
 ┣ 📜file-structure.md
 ┣ 📜go.mod
 ┣ 📜go.sum
 ┗ 📜readme.md