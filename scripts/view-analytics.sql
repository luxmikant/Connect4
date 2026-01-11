-- Quick Analytics Queries for Connect4 Multiplayer

-- ========================================
-- ANALYTICS SNAPSHOTS
-- ========================================

-- Latest Analytics Snapshot
SELECT 
    timestamp,
    games_completed_hour as "Games/Hour",
    games_completed_day as "Games/Day",
    avg_game_duration_sec as "Avg Duration (sec)",
    min_game_duration_sec as "Min Duration",
    max_game_duration_sec as "Max Duration",
    unique_players_hour as "Unique Players/Hour",
    active_games as "Active Games"
FROM analytics_snapshots
ORDER BY timestamp DESC
LIMIT 1;

-- Analytics Over Time (Last 24 Hours)
SELECT 
    DATE_TRUNC('hour', timestamp) as hour,
    AVG(games_completed_hour)::INTEGER as avg_games_per_hour,
    AVG(avg_game_duration_sec)::INTEGER as avg_duration_sec,
    AVG(unique_players_hour)::INTEGER as avg_unique_players
FROM analytics_snapshots
WHERE timestamp >= NOW() - INTERVAL '24 hours'
GROUP BY hour
ORDER BY hour DESC;

-- ========================================
-- PLAYER STATISTICS
-- ========================================

-- Top 10 Players by Wins
SELECT 
    username,
    games_won as "Wins",
    games_played as "Games",
    ROUND(win_rate * 100, 2) as "Win Rate %",
    ROUND(avg_game_time) as "Avg Game Time (sec)"
FROM player_stats
WHERE games_played > 0
ORDER BY games_won DESC
LIMIT 10;

-- Top 10 Players by Win Rate (min 3 games)
SELECT 
    username,
    games_won as "Wins",
    games_played as "Games",
    ROUND(win_rate * 100, 2) as "Win Rate %"
FROM player_stats
WHERE games_played >= 3
ORDER BY win_rate DESC, games_won DESC
LIMIT 10;

-- ========================================
-- GAME STATISTICS
-- ========================================

-- Recent Games
SELECT 
    id,
    player1,
    player2,
    status,
    winner,
    EXTRACT(EPOCH FROM (COALESCE(end_time, NOW()) - start_time))::INTEGER as duration_sec,
    created_at
FROM game_sessions
ORDER BY created_at DESC
LIMIT 10;

-- Games Completed Today
SELECT 
    COUNT(*) as total_games,
    AVG(EXTRACT(EPOCH FROM (end_time - start_time)))::INTEGER as avg_duration_sec,
    COUNT(DISTINCT player1) + COUNT(DISTINCT player2) as unique_players
FROM game_sessions
WHERE status = 'completed'
  AND created_at >= CURRENT_DATE;

-- Games by Hour of Day
SELECT 
    EXTRACT(HOUR FROM created_at) as hour_of_day,
    COUNT(*) as game_count
FROM game_sessions
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY hour_of_day
ORDER BY game_count DESC;

-- ========================================
-- GAME EVENTS
-- ========================================

-- Recent Game Events
SELECT 
    event_type,
    game_id,
    player_id,
    timestamp,
    metadata
FROM game_events
ORDER BY timestamp DESC
LIMIT 20;

-- Event Counts by Type
SELECT 
    event_type,
    COUNT(*) as event_count
FROM game_events
WHERE timestamp >= NOW() - INTERVAL '24 hours'
GROUP BY event_type
ORDER BY event_count DESC;

-- ========================================
-- ACTIVE GAMES
-- ========================================

-- Currently Active Games
SELECT 
    id,
    player1,
    player2,
    current_turn,
    EXTRACT(EPOCH FROM (NOW() - start_time))::INTEGER as duration_so_far_sec,
    start_time
FROM game_sessions
WHERE status = 'in_progress'
ORDER BY start_time DESC;

-- ========================================
-- SUMMARY DASHBOARD
-- ========================================

-- Overall System Summary
SELECT 
    'Total Games' as metric,
    COUNT(*)::TEXT as value
FROM game_sessions
UNION ALL
SELECT 
    'Completed Games',
    COUNT(*)::TEXT
FROM game_sessions
WHERE status = 'completed'
UNION ALL
SELECT 
    'Active Games',
    COUNT(*)::TEXT
FROM game_sessions
WHERE status = 'in_progress'
UNION ALL
SELECT 
    'Total Players',
    COUNT(DISTINCT username)::TEXT
FROM player_stats
UNION ALL
SELECT 
    'Games Today',
    COUNT(*)::TEXT
FROM game_sessions
WHERE created_at >= CURRENT_DATE
UNION ALL
SELECT 
    'Avg Game Duration',
    ROUND(AVG(avg_game_time))::TEXT || ' sec'
FROM player_stats
WHERE games_played > 0;

-- ========================================
-- KAFKA EVENT MONITORING
-- ========================================

-- Latest Events Published to Kafka (from game_events table)
SELECT 
    event_type,
    game_id,
    player_id,
    timestamp,
    metadata->>'winner' as winner,
    metadata->>'duration' as duration_sec
FROM game_events
ORDER BY timestamp DESC
LIMIT 25;

-- Events Per Hour (Last 24h)
SELECT 
    DATE_TRUNC('hour', timestamp) as hour,
    event_type,
    COUNT(*) as event_count
FROM game_events
WHERE timestamp >= NOW() - INTERVAL '24 hours'
GROUP BY hour, event_type
ORDER BY hour DESC, event_count DESC;
