#!/bin/bash

# Test script for Tashi Backend API
# Make sure the server is running before executing this script

BASE_URL="http://localhost:8080"

echo "Testing Tashi Backend API..."
echo "=============================="

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s "$BASE_URL/health" | jq '.' || echo "Health endpoint failed"
echo ""

# Test system info endpoint
echo "2. Testing system info endpoint..."
curl -s "$BASE_URL/api/system/info" | jq '.' || echo "System info endpoint failed"
echo ""

# Test templates endpoint
echo "3. Testing templates endpoint..."
curl -s "$BASE_URL/api/templates" | jq '.' || echo "Templates endpoint failed"
echo ""

# Test creating a task
echo "4. Testing task creation..."
TASK_RESPONSE=$(curl -s -X POST "$BASE_URL/api/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Create a simple script that prints Hello World",
    "platform": "linux",
    "language": "bash"
  }')

echo "$TASK_RESPONSE" | jq '.' || echo "Task creation failed"

# Extract task ID
TASK_ID=$(echo "$TASK_RESPONSE" | jq -r '.data.id')
echo "Created task with ID: $TASK_ID"
echo ""

# Test getting task status
echo "5. Testing task status retrieval..."
curl -s "$BASE_URL/api/tasks/$TASK_ID" | jq '.' || echo "Task status retrieval failed"
echo ""

# Test getting all tasks
echo "6. Testing get all tasks..."
curl -s "$BASE_URL/api/tasks" | jq '.' || echo "Get all tasks failed"
echo ""

# Test stats endpoint
echo "7. Testing stats endpoint..."
curl -s "$BASE_URL/api/stats" | jq '.' || echo "Stats endpoint failed"
echo ""

echo "API testing completed!"
echo "Note: Task processing is async, so the task may still be processing."
echo "Check the task status again after a few seconds to see the generated script."
