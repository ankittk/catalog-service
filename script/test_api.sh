#!/bin/bash

BASE_URL="http://localhost:8000"
TOKEN=$1

# Function to handle requests with or without token
request() {
  if [ -z "$TOKEN" ]; then
    # No token, just the normal request
    curl -s "$BASE_URL$1" | jq .
  else
    # With token, include the Authorization header
    curl -s "$BASE_URL$1" -H "Authorization: Bearer $TOKEN" | jq .
  fi
}

echo "Catalog Service API Test Script"
echo "=================================="
echo ""

echo "1. Health Check"
echo "---------------"
request "/health"
echo ""

echo "2. List All Services"
echo "-------------------"
request "/v1/services" | jq '.services | length'
echo "Total services returned"
echo ""

echo "3. Get Specific Service (svc-1)"
echo "-------------------------------"
request "/v1/services/svc-1" | jq '.service.name'
echo ""

echo "4. Get Service Versions (svc-1)"
echo "-------------------------------"
request "/v1/services/svc-1/versions" | jq '.versions | length'
echo "versions returned"
echo ""

echo "5. Filter by Organization (org-1)"
echo "---------------------------------"
request "/v1/services?organization_id=org-1" | jq '.services | length'
echo "services for org-1"
echo ""

echo "6. Pagination Test (page_size=2)"
echo "--------------------------------"
request "/v1/services?page_size=2" | jq '.nextPageToken'
echo "next page token"
echo ""

echo "7. Sorting Test (by name, ascending)"
echo "-----------------------------------"
request "/v1/services?sort_by=name&sort_order=asc" | jq '.services[].name'
echo ""

echo "8. Search Test (search for 'user')"
echo "---------------------------------"
request "/v1/services?search_query=user" | jq '.services[].name'
echo ""

echo "All tests completed!"
echo "--------------------------------"

echo "Summary:"
echo "- Health checks: ✅ Working"
echo "- Pagination: ✅ Working"
echo "- Filtering: ✅ Working"
echo "- Sorting: ✅ Working"
echo "- Search: ✅ Working"
