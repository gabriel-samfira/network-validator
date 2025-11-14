#!/bin/bash

# Test script to demonstrate HTTP method routing with Go 1.22+ enhanced patterns
# Usage: ./test_http_methods.sh

BASE_URL="http://localhost:8080"

echo "🧪 Testing HTTP Method Routing with Go 1.22+ Enhanced Patterns"
echo "============================================================="
echo ""

# Test GET methods
echo "📥 Testing GET requests:"
echo "------------------------"

echo "1. GET /api/sysinfo (basic system info):"
curl -s -X GET "$BASE_URL/api/sysinfo" | jq '.hostname, .os.name' 2>/dev/null || curl -s -X GET "$BASE_URL/api/sysinfo"
echo ""

echo "2. GET /api/health:"
curl -s -X GET "$BASE_URL/api/health" | jq '.status' 2>/dev/null || curl -s -X GET "$BASE_URL/api/health"
echo ""

echo "3. GET /api/config:"
curl -s -X GET "$BASE_URL/api/config" | jq '.' 2>/dev/null || curl -s -X GET "$BASE_URL/api/config"
echo ""

# Test POST methods
echo "📤 Testing POST requests:"
echo "-------------------------"

echo "4. POST /api/sysinfo (filtered system info):"
curl -s -X POST "$BASE_URL/api/sysinfo" \
  -H "Content-Type: application/json" \
  -d '{"fields": ["hostname", "cpu"], "format": "json"}' | jq '.' 2>/dev/null || curl -s -X POST "$BASE_URL/api/sysinfo" \
  -H "Content-Type: application/json" \
  -d '{"fields": ["hostname", "cpu"], "format": "json"}'
echo ""

echo "5. POST /api/health (health check with parameters):"
curl -s -X POST "$BASE_URL/api/health" \
  -H "Content-Type: application/json" \
  -d '{"check_type": "detailed"}' | jq '.' 2>/dev/null || curl -s -X POST "$BASE_URL/api/health" \
  -H "Content-Type: application/json" \
  -d '{"check_type": "detailed"}'
echo ""

echo "6. POST /api/config (create new configuration):"
curl -s -X POST "$BASE_URL/api/config" \
  -H "Content-Type: application/json" \
  -d '{"log_level": "debug", "refresh_rate": 60, "enable_cors": true}' | jq '.' 2>/dev/null || curl -s -X POST "$BASE_URL/api/config" \
  -H "Content-Type: application/json" \
  -d '{"log_level": "debug", "refresh_rate": 60, "enable_cors": true}'
echo ""

# Test PUT method
echo "🔄 Testing PUT request:"
echo "-----------------------"

echo "7. PUT /api/config (update configuration):"
curl -s -X PUT "$BASE_URL/api/config" \
  -H "Content-Type: application/json" \
  -d '{"log_level": "warn", "refresh_rate": 45}' | jq '.' 2>/dev/null || curl -s -X PUT "$BASE_URL/api/config" \
  -H "Content-Type: application/json" \
  -d '{"log_level": "warn", "refresh_rate": 45}'
echo ""

# Test DELETE method
echo "🗑️  Testing DELETE request:"
echo "---------------------------"

echo "8. DELETE /api/config (reset to defaults):"
curl -s -X DELETE "$BASE_URL/api/config" | jq '.' 2>/dev/null || curl -s -X DELETE "$BASE_URL/api/config"
echo ""

# Test method not allowed
echo "❌ Testing unsupported method:"
echo "------------------------------"

echo "9. PATCH /api/sysinfo (should return 405 Method Not Allowed):"
curl -s -X PATCH "$BASE_URL/api/sysinfo" | jq '.' 2>/dev/null || curl -s -X PATCH "$BASE_URL/api/sysinfo"
echo ""

echo "✅ HTTP Method Routing Test Complete!"
echo ""
echo "💡 Key improvements with Go 1.22+ enhanced routing:"
echo "   • Cleaner syntax: mux.HandleFunc(\"GET /api/health\", handler)"
echo "   • Built-in method matching (no custom methodHandler needed)"
echo "   • Automatic 405 Method Not Allowed responses"
echo "   • Better performance (no runtime method checking)"
echo "   • Pattern matching with path parameters support"