# Kafka Cloud Testing Guide

## Overview

This guide provides multiple ways to test and verify your Kafka Cloud (Confluent Cloud) connection is working properly.

## Quick Validation

### 1. Run Configuration Check
```bash
# Validate your Kafka setup
powershell -ExecutionPolicy Bypass -File scripts/validate-kafka-setup.ps1
```

This script will:
- ✅ Check all Kafka environment variables
- ✅ Test compilation of producer/consumer
- ✅ Provide next steps for testing

## Testing Methods

### Method 1: Command Line Producer Test

**Send test messages to Kafka:**
```bash
go run scripts/test-kafka-cloud.go
```

**What it tests:**
- ✅ Producer creation and configuration
- ✅ Authentication with Confluent Cloud
- ✅ Sending different event types
- ✅ Performance (10 rapid messages)
- ✅ Error handling and timeouts

**Expected output:**
```
🔍 Testing Kafka Cloud Connection (Confluent Cloud)
============================================================
📋 Configuration:
   Bootstrap Servers: pkc-9q8rv.ap-south-2.aws.confluent.cloud:9092
   Topic: game-events
   Consumer Group: analytics-service
   API Key: MKMWUHNR...
   API Secret: cfltX7sO...

🔧 Test 1: Creating Kafka Producer...
✅ Producer created successfully

📤 Test 2: Sending Test Events...
   1. Sending Player Joined Event... ✅ SUCCESS (45ms)
   2. Sending Game Started Event... ✅ SUCCESS (32ms)
   3. Sending Move Event... ✅ SUCCESS (28ms)
   4. Sending Game Completed Event... ✅ SUCCESS (31ms)

⚡ Test 3: Performance Test (10 rapid events)...
✅ Sent 10 events in 287ms (avg: 28ms per event)

📊 Test Summary:
============================================================
✅ Producer Creation: SUCCESS
✅ Events Sent: 4/4
✅ Performance: 28ms avg per event

🎉 ALL TESTS PASSED - Kafka Cloud is working perfectly!
```

### Method 2: Consumer Test (Analytics Service)

**Test message consumption:**
```bash
go run scripts/test-kafka-consumer.go
```

**What it tests:**
- ✅ Consumer creation and configuration
- ✅ Database connection for analytics
- ✅ Message processing and storage
- ✅ Real-time message consumption

**Usage:**
1. Start the consumer in one terminal
2. Run the producer test in another terminal
3. Watch messages being processed in real-time

### Method 3: Web Monitor Interface

**Start the web monitor:**
```bash
go run scripts/kafka-monitor.go
```

**Then open:** http://localhost:8081

**Features:**
- 🖥️ Web interface for monitoring
- 📤 Send test messages with buttons
- 📊 View sent messages in real-time
- 🎯 Easy testing without command line

### Method 4: Confluent Cloud Console

**Check messages in Confluent Cloud:**

1. **Login to Confluent Cloud:**
   - Go to: https://confluent.cloud/
   - Login with your account

2. **Navigate to your cluster:**
   - Environments → Your Environment
   - Clusters → Your Cluster (pkc-9q8rv)

3. **Check the topic:**
   - Topics → `game-events`
   - Messages tab
   - You should see test messages appearing

4. **Monitor metrics:**
   - Overview tab shows throughput
   - Metrics tab shows detailed statistics

## Troubleshooting

### Common Issues

**1. Authentication Errors**
```
Error: SASL authentication failed
```
**Solution:** Check your API key and secret in `.env`

**2. Network Timeouts**
```
Error: context deadline exceeded
```
**Solution:** Check your internet connection and firewall

**3. Topic Not Found**
```
Error: topic does not exist
```
**Solution:** Create the topic in Confluent Cloud console

**4. Permission Denied**
```
Error: not authorized to access topic
```
**Solution:** Check API key permissions in Confluent Cloud

### Debug Steps

1. **Verify Configuration:**
   ```bash
   powershell -ExecutionPolicy Bypass -File scripts/validate-kafka-setup.ps1
   ```

2. **Test Network Connectivity:**
   ```bash
   # Test if you can reach Confluent Cloud
   nslookup pkc-9q8rv.ap-south-2.aws.confluent.cloud
   ```

3. **Check Confluent Cloud Status:**
   - Visit: https://status.confluent.io/
   - Ensure no ongoing incidents

4. **Verify API Key Permissions:**
   - Confluent Cloud Console → API Keys
   - Ensure key has produce/consume permissions

## Performance Benchmarks

### Expected Performance
- **Message Send Time:** < 50ms per message
- **Batch Send Time:** < 30ms per message average
- **Connection Setup:** < 2 seconds
- **Authentication:** < 1 second

### Performance Test Results
```bash
# Run performance test
go run scripts/test-kafka-cloud.go

# Look for these metrics:
✅ Sent 10 events in 287ms (avg: 28ms per event)
```

**Good Performance:** < 50ms average
**Excellent Performance:** < 30ms average

## Integration Testing

### Full System Test

1. **Start Analytics Service:**
   ```bash
   go run cmd/analytics/main.go
   ```

2. **Send Test Events:**
   ```bash
   go run scripts/test-kafka-cloud.go
   ```

3. **Check Database:**
   ```sql
   -- Connect to your Supabase database
   SELECT * FROM game_events ORDER BY created_at DESC LIMIT 10;
   ```

4. **Verify Processing:**
   - Events should appear in database
   - Analytics service should log processing
   - No errors in either service

## Monitoring in Production

### Key Metrics to Watch

1. **Message Throughput:**
   - Messages per second
   - Batch sizes

2. **Latency:**
   - Producer send time
   - Consumer processing time

3. **Error Rates:**
   - Failed sends
   - Processing errors

4. **Consumer Lag:**
   - How far behind consumers are
   - Processing backlog

### Confluent Cloud Monitoring

**Built-in Dashboards:**
- Cluster Overview → Metrics
- Topic-level metrics
- Consumer group lag
- Throughput and latency graphs

**Alerts:**
- Set up alerts for high latency
- Monitor consumer lag
- Track error rates

## Next Steps

Once Kafka testing is complete:

1. ✅ **Kafka Working** → Proceed with REST API development
2. ✅ **Analytics Ready** → Integrate event tracking in game logic
3. ✅ **Monitoring Setup** → Use web monitor during development
4. ✅ **Production Ready** → Deploy with confidence

Your Kafka Cloud setup is now fully tested and ready for production use!