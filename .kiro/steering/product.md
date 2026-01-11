# Product Overview

## Connect 4 Multiplayer Game System

A real-time multiplayer Connect 4 game platform supporting player-vs-player and player-vs-bot gameplay with comprehensive analytics tracking.

### Core Features
- **Real-time multiplayer gameplay** via WebSocket connections
- **Intelligent bot opponents** using minimax algorithm with alpha-beta pruning
- **Automatic matchmaking** with 10-second timeout fallback to bot games
- **Player reconnection support** with 30-second session persistence
- **Live leaderboard** with player statistics and rankings
- **Analytics pipeline** for game metrics and player behavior tracking

### Target Users
- Casual gamers seeking quick Connect 4 matches
- Competitive players wanting ranked gameplay
- Single players practicing against AI opponents
- Product managers analyzing game engagement metrics

### Success Metrics
- Average matchmaking time under 10 seconds
- Bot response time under 1 second
- Real-time updates delivered within 100ms
- Player retention and engagement analytics

## AI Assistant Guidelines

### Product Context for Development
When working on this project, always consider:

**User Experience Priorities:**
1. **Speed**: Fast matchmaking and responsive gameplay
2. **Reliability**: Stable connections and session recovery
3. **Engagement**: Challenging bot AI and competitive features
4. **Analytics**: Comprehensive tracking for product insights

**Business Logic Rules:**
- Games must start within 10 seconds (matchmaking timeout)
- Bot moves must complete within 1 second
- WebSocket updates must be under 100ms latency
- Sessions persist for 30 seconds after disconnection
- Leaderboard updates in real-time after each game

**Quality Standards:**
- All game logic must be deterministic and testable
- Real-time features require comprehensive error handling
- Analytics events must be reliable and complete
- User inputs must be validated and sanitized

### Feature Development Approach
- **Start with core game mechanics** before adding advanced features
- **Implement bot AI early** as it's critical for single-player experience
- **Build WebSocket infrastructure** before frontend integration
- **Add analytics throughout** rather than retrofitting later