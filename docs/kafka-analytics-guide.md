# Kafka Analytics Guide

## Overview

The Connect4 multiplayer game includes **real-time analytics powered by Apache Kafka**. Game events are published to Kafka topics and consumed by an analytics service that tracks gameplay metrics.

## Architecture

```
Game Server → Kafka Producer → Kafka Topic (game-events)
                                      ↓
                              Kafka Consumer (Analytics Service)
                                      ↓
                              PostgreSQL (analytics_snapshots)
```

## Analytics Service Location

The analytics implementation is located in:
- **Service**: `internal/analytics/service.go`
- **Producer**: `internal/analytics/producer.go`
- **Consumer (Main)**: `cmd/analytics/main.go`
- **Database Migration**: `migrations/007_create_analytics_snapshots_table.sql`

## Tracked Metrics

### Gameplay Metrics
- ✅ **Average game duration** - Mean time to complete games
- ✅ **Games per hour** - Games completed in the last hour
- ✅ **Games per day** - Games completed in the last 24 hours
- ✅ **Min/Max game duration** - Shortest and longest games
- ✅ **Total moves** - Aggregate move count

### User Metrics
- ✅ **Unique players per hour** - Distinct players active
- ✅ **Active games** - Currently in-progress games
- ✅ **Player statistics** - Individual win rates and game counts (in `player_stats` table)

## How to Run Analytics Service

### 1. Start the Analytics Consumer

In a separate terminal:

```bash
# Build the analytics service
go build -o analytics.exe ./cmd/analytics

# Run the analytics consumer
./analytics.exe
```

**Expected Output:**
```
Starting analytics service...
Analytics service initialized
Kafka consumer group: analytics-service
Topic: game-events
Listening for game events...
```

### 2. The Service Will Automatically:

- Listen to Kafka topic `game-events`
- Process events: `game_started`, `move_made`, `game_completed`
- Calculate real-time metrics
- Store snapshots every 5 minutes in `analytics_snapshots` table

## How to View Analytics Data

### Option 1: Direct Database Query

Connect to PostgreSQL and run:

```sql
-- View latest analytics snapshot
SELECT 
    timestamp,
    games_completed_hour,
    games_completed_day,
    avg_game_duration_sec,
    unique_players_hour,
    active_games
FROM analytics_snapshots
ORDER BY timestamp DESC
LIMIT 10;
```

### Option 2: Query Specific Metrics

```sql
-- Average game duration over time
SELECT 
    DATE_TRUNC('hour', timestamp) as hour,
    AVG(avg_game_duration_sec) as avg_duration,
    SUM(games_completed_hour) as total_games
FROM analytics_snapshots
GROUP BY hour
ORDER BY hour DESC;

-- Peak gaming hours
SELECT 
    EXTRACT(HOUR FROM timestamp) as hour_of_day,
    AVG(games_completed_hour) as avg_games,
    MAX(unique_players_hour) as max_players
FROM analytics_snapshots
GROUP BY hour_of_day
ORDER BY avg_games DESC;
```

### Option 3: Player Statistics

```sql
-- Most frequent winners
SELECT 
    username,
    games_won,
    games_played,
    win_rate,
    ROUND(avg_game_time) as avg_game_time_sec
FROM player_stats
ORDER BY games_won DESC
LIMIT 10;

-- Players by win rate (min 3 games)
SELECT 
    username,
    games_won,
    games_played,
    ROUND(win_rate * 100, 2) as win_rate_pct
FROM player_stats
WHERE games_played >= 3
ORDER BY win_rate DESC
LIMIT 10;
```

## Kafka Configuration

The analytics service uses the Kafka configuration from `config.yaml`:

```yaml
kafka:
  bootstrap_servers: "pkc-9q8rv.ap-south-2.aws.confluent.cloud:9092"
  api_key: "YOUR_API_KEY"
  api_secret: "YOUR_API_SECRET"
  topic: "game-events"
  consumer_group: "analytics-service"
```

## Event Types Published

### 1. game_started
```json
{
  "event_type": "game_started",
  "game_id": "uuid",
  "player1": "username1",
  "player2": "username2",
  "timestamp": "2026-01-05T12:00:00Z"
}
```

### 2. move_made
```json
{
  "event_type": "move_made",
  "game_id": "uuid",
  "player": "username",
  "column": 3,
  "timestamp": "2026-01-05T12:01:00Z"
}
```

### 3. game_completed
```json
{
  "event_type": "game_completed",
  "game_id": "uuid",
  "winner": "username",
  "duration": 120,
  "timestamp": "2026-01-05T12:02:00Z"
}
```

## Monitoring Kafka Events

### Using Kafka CLI Tools

```bash
# List topics
kafka-topics --bootstrap-server pkc-9q8rv.ap-south-2.aws.confluent.cloud:9092 \
  --command-config client.properties \
  --list

# Consume events from beginning
kafka-console-consumer --bootstrap-server pkc-9q8rv.ap-south-2.aws.confluent.cloud:9092 \
  --consumer.config client.properties \
  --topic game-events \
  --from-beginning
```

### Using Confluent Cloud Dashboard

1. Login to https://confluent.cloud
2. Navigate to your cluster
3. Click "Topics" → "game-events"
4. View messages, metrics, and consumer lag

## Testing Analytics

### 1. Play Some Games

```bash
# Start game server
./server.exe

# Open browser: http://localhost:5173
# Play 3-5 games (vs bot or vs player)
```

### 2. Check Events Were Published

```bash
# Start analytics consumer
./analytics.exe

# You should see logs like:
# Processed game_started event: game_id=...
# Processed move_made event: game_id=...
# Processed game_completed event: winner=...
# Snapshot created: games_completed=5, avg_duration=45s
```

### 3. Query Analytics Data

```sql
-- Check if snapshots are being created
SELECT COUNT(*) FROM analytics_snapshots;

-- View latest metrics
SELECT * FROM analytics_snapshots ORDER BY timestamp DESC LIMIT 1;
```

## Troubleshooting

### No Events Being Consumed

1. **Check Kafka connection:**
   ```bash
   # Verify credentials in config.yaml
   # Test connectivity to Confluent Cloud
   ```

2. **Check topic exists:**
   - Login to Confluent Cloud
   - Verify `game-events` topic is created

3. **Check consumer group:**
   ```bash
   # View consumer lag in Confluent Cloud dashboard
   ```

### Analytics Service Crashes

1. **Check database connection:**
   ```sql
   -- Verify table exists
   \dt analytics_snapshots
   ```

2. **Check logs:**
   ```bash
   # Analytics service outputs detailed error logs
   ./analytics.exe 2>&1 | tee analytics.log
   ```

## Production Considerations

### Scaling
- Run multiple analytics service instances for high throughput
- Kafka consumer group handles automatic partition distribution

### Retention
- Configure snapshot retention policy:
  ```sql
  -- Delete snapshots older than 90 days
  DELETE FROM analytics_snapshots 
  WHERE timestamp < NOW() - INTERVAL '90 days';
  ```

### Monitoring
- Track consumer lag via Confluent Cloud
- Set up alerts for processing delays
- Monitor database storage for snapshots table

## Additional Resources

- [Kafka Testing Guide](kafka-testing-guide.md)
- [Kafka Windows Setup](kafka-windows-troubleshooting.md)
- [Integration Testing](integration-testing.md)
