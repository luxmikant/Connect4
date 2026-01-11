# Cloud Infrastructure Alternatives (No Docker)

## Overview

This document provides cloud-based alternatives to the Docker Compose setup, utilizing your Confluent Cloud account and cloud database services for development and production deployment.

## 🔍 Redis Usage Analysis

**Current Redis Usage in Our Design:**
- **Purpose**: In-memory storage for active game sessions
- **Type**: **Non-functional requirement** (performance optimization)
- **Core Function**: Session state caching and fast lookups
- **Alternative**: Can be replaced with PostgreSQL + application-level caching

**Redis Functions in Connect 4:**
1. **Active Game Sessions**: Store current game state for fast access
2. **Player Connection Mapping**: Map WebSocket connections to players
3. **Matchmaking Queue**: Temporary storage for players waiting for games
4. **Session Timeouts**: Handle 30-second reconnection windows

**Verdict**: Redis is **NOT essential** - it's a performance optimization that can be replaced.

## 🌐 Cloud Infrastructure Setup

### 1. PostgreSQL - Cloud Database Options

#### Option A: Supabase (Recommended for Development)
```yaml
# Free tier: 500MB database, 2 concurrent connections
# Pros: Easy setup, built-in auth, real-time subscriptions
# Cons: Limited free tier

Database URL: postgresql://[user]:[password]@db.[project].supabase.co:5432/postgres
```

**Setup Steps:**
1. Go to [supabase.com](https://supabase.com)
2. Create new project
3. Get connection string from Settings → Database
4. Use in your Go application

#### Option B: Neon (Serverless PostgreSQL)
```yaml
# Free tier: 3GB storage, 1 database
# Pros: Serverless, auto-scaling, generous free tier
# Cons: Cold starts

Database URL: postgresql://[user]:[password]@[endpoint].neon.tech/[dbname]
```

#### Option C: Railway
```yaml
# Free tier: $5 credit monthly
# Pros: Simple deployment, good for development
# Cons: Credit-based pricing

Database URL: postgresql://[user]:[password]@[host].railway.app:5432/railway
```

#### Option D: AWS RDS Free Tier
```yaml
# Free tier: 20GB storage, 750 hours/month
# Pros: Production-ready, AWS ecosystem
# Cons: More complex setup

Database URL: postgresql://[user]:[password]@[endpoint].rds.amazonaws.com:5432/connect4
```

### 2. Kafka - Confluent Cloud (You Already Have This!)

Since you have Confluent Cloud Dev Pro account:

```yaml
# Confluent Cloud Configuration
Bootstrap Servers: [your-cluster].confluent.cloud:9092
Security Protocol: SASL_SSL
SASL Mechanism: PLAIN
SASL Username: [your-api-key]
SASL Password: [your-api-secret]

# Topics to create:
- game-events
- player-events
- analytics-events
```

**Go Configuration:**
```go
config := &kafka.ConfigMap{
    "bootstrap.servers": "your-cluster.confluent.cloud:9092",
    "security.protocol": "SASL_SSL",
    "sasl.mechanisms":   "PLAIN",
    "sasl.username":     "your-api-key",
    "sasl.password":     "your-api-secret",
}
```

### 3. Redis Alternatives (In-Memory Storage)

#### Option A: No Redis - PostgreSQL Only (Simplest)
```go
// Store active sessions directly in PostgreSQL
type GameSession struct {
    ID          string    `gorm:"primaryKey"`
    Player1ID   string    
    Player2ID   string    
    BoardState  string    `gorm:"type:jsonb"` // JSON storage
    Status      string    
    LastActivity time.Time `gorm:"index"`     // For cleanup
    // ... other fields
}

// Fast queries with proper indexing
func (r *GameRepository) GetActiveSession(playerID string) (*GameSession, error) {
    var session GameSession
    err := r.db.Where("(player1_id = ? OR player2_id = ?) AND status = ?", 
        playerID, playerID, "active").First(&session).Error
    return &session, err
}
```

**Pros**: Simple, no additional infrastructure, ACID guarantees
**Cons**: Slightly slower than Redis (but likely unnoticeable for Connect 4)

#### Option B: Application-Level In-Memory Cache
```go
// Simple in-memory cache with sync.Map
type SessionCache struct {
    sessions sync.Map // map[string]*GameSession
    db       *gorm.DB
}

func (c *SessionCache) GetSession(id string) (*GameSession, error) {
    // Try cache first
    if session, ok := c.sessions.Load(id); ok {
        return session.(*GameSession), nil
    }
    
    // Fallback to database
    session, err := c.loadFromDB(id)
    if err == nil {
        c.sessions.Store(id, session)
    }
    return session, err
}
```

**Pros**: Fast access, no external dependencies
**Cons**: Data lost on restart, no sharing between instances

#### Option C: Cloud Redis (If You Want Redis)

**Redis Cloud (Free Tier):**
- 30MB free tier
- Managed Redis service
- Connection string: `redis://:[password]@[endpoint]:port`

**AWS ElastiCache (Free Tier):**
- 750 hours/month free
- Production-ready
- VPC integration

## 🚀 Recommended Setup for Your Project

### Development Environment
```yaml
Database: Supabase (Free tier, easy setup)
Kafka: Your Confluent Cloud account
In-Memory: PostgreSQL only (no Redis needed)
```

### Production Environment
```yaml
Database: Neon or AWS RDS
Kafka: Your Confluent Cloud account  
In-Memory: Application-level cache or Redis Cloud
```

## 📝 Updated Configuration

### Environment Variables
```bash
# Database
DATABASE_URL=postgresql://user:pass@host:5432/dbname

# Kafka (Confluent Cloud)
KAFKA_BOOTSTRAP_SERVERS=your-cluster.confluent.cloud:9092
KAFKA_API_KEY=your-api-key
KAFKA_API_SECRET=your-api-secret

# Redis (Optional)
REDIS_URL=redis://user:pass@host:port

# Application
PORT=8080
ENV=development
```

### Go Configuration Structure
```go
type Config struct {
    Database DatabaseConfig `mapstructure:"database"`
    Kafka    KafkaConfig    `mapstructure:"kafka"`
    Redis    *RedisConfig   `mapstructure:"redis,omitempty"` // Optional
    Server   ServerConfig   `mapstructure:"server"`
}

type DatabaseConfig struct {
    URL             string `mapstructure:"url"`
    MaxOpenConns    int    `mapstructure:"max_open_conns"`
    MaxIdleConns    int    `mapstructure:"max_idle_conns"`
    ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
}

type KafkaConfig struct {
    BootstrapServers string `mapstructure:"bootstrap_servers"`
    APIKey          string `mapstructure:"api_key"`
    APISecret       string `mapstructure:"api_secret"`
    TopicPrefix     string `mapstructure:"topic_prefix"`
}

type RedisConfig struct {
    URL        string `mapstructure:"url"`
    MaxRetries int    `mapstructure:"max_retries"`
}
```

## 🔧 Implementation Changes

### Updated Task 1: Project Setup
```markdown
- [ ] 1. Project Setup and Core Infrastructure
  - Initialize Go modules and project structure
  - Set up cloud database connection (Supabase/Neon)
  - Configure Confluent Cloud Kafka integration
  - Implement PostgreSQL-based session storage (no Redis)
  - Set up configuration management with Viper
  - Create Makefile with cloud-based development commands
```

### Session Management Without Redis
```go
// internal/session/manager.go
type Manager struct {
    db    *gorm.DB
    cache sync.Map // Optional in-memory cache
}

func (m *Manager) CreateSession(player1, player2 string) (*GameSession, error) {
    session := &GameSession{
        ID:        uuid.New().String(),
        Player1ID: player1,
        Player2ID: player2,
        Status:    "active",
        CreatedAt: time.Now(),
    }
    
    // Save to PostgreSQL
    if err := m.db.Create(session).Error; err != nil {
        return nil, err
    }
    
    // Cache in memory (optional)
    m.cache.Store(session.ID, session)
    
    return session, nil
}

func (m *Manager) GetActiveSession(playerID string) (*GameSession, error) {
    var session GameSession
    err := m.db.Where(
        "(player1_id = ? OR player2_id = ?) AND status = ?", 
        playerID, playerID, "active",
    ).First(&session).Error
    
    return &session, err
}
```

## 💰 Cost Analysis

### Free Tier Limits
```yaml
Supabase: 500MB database, 2 concurrent connections
Confluent Cloud: Your existing Dev Pro account
Total Monthly Cost: $0 for development
```

### Production Scaling
```yaml
Neon: ~$20/month for production database
Confluent Cloud: Your existing account
Redis Cloud: $0-5/month for small cache
Total: ~$20-25/month
```

## ✅ Benefits of This Approach

1. **No Docker Complexity**: Direct cloud connections
2. **Cost Effective**: Free tiers for development
3. **Production Ready**: Same services scale to production
4. **Simplified Architecture**: Fewer moving parts
5. **Your Existing Assets**: Leverages your Confluent account

## 🚀 Next Steps

1. **Choose Database**: Recommend Supabase for quick start
2. **Set up Confluent Topics**: Create game-events, player-events topics
3. **Skip Redis**: Use PostgreSQL + optional in-memory cache
4. **Update Task 1**: Focus on cloud connections instead of Docker

Would you like me to update the tasks.md file with these cloud-based alternatives?