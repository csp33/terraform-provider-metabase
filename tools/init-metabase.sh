#!/bin/bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


# 1. Get the setup-token
echo "Getting the setup-token..."
response_properties=$(curl --location 'http://localhost:3000/api/session/properties' -s)

if [ $? -ne 0 ]; then
  echo "Error getting session properties."
  exit 1
fi

setup_token=$(echo "$response_properties" | jq -r '.setup-token')

if [ -z "$setup_token" ]; then
  echo "Could not extract the setup-token."
  exit 1
fi

echo "Setup-token obtained: $setup_token"

# 2. Perform the setup
echo "Performing the setup..."
setup_response=$(curl --location 'http://localhost:3000/api/setup' \
  --header 'Content-Type: application/json' \
  --data-raw "{
      \"user\": {
          \"email\": \"test@test.com\",
          \"password\": \"testpwd1\"
      },
      \"token\": \"$setup_token\",
      \"prefs\": {
          \"site_name\": \"metabase\"
      }
  }" -s)

if [ $? -ne 0 ]; then
  echo "Error performing the setup."
  echo "Details: $setup_response"
  exit 1
fi

session_token=$(echo "$setup_response" | jq -r '.token')

if [ -z "$session_token" ]; then
  echo "Could not extract the session token from the setup."
  echo "Setup response: $setup_response"
  exit 1
fi

echo "Setup completed. Session token obtained: $session_token"

# 3. Create the API Key
echo "Creating the API Key..."
api_key_response=$(curl --location 'http://localhost:3000/api/api-key' \
  --header "X-Metabase-Session: $session_token" \
  --header 'Content-Type: application/json' \
  --data '{
    "group_id": 2,
    "name": "API Key 2"
  }' -s)

if [ $? -ne 0 ]; then
  echo "Error creating the API Key."
  echo "Details: $api_key_response"
  exit 1
fi

api_key=$(echo "$api_key_response" | jq -r '.unmasked_key')

if [ -z "$api_key" ]; then
  echo "Could not extract the API Key."
  echo "API Key creation response: $api_key_response"
  exit 1
fi


echo $api_key
