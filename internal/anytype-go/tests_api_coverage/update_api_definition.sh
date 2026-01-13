#!/bin/bash
# This script is used to update the API definition file with the latest API endpoints.
# It will fetch the latest API endpoints from the server and update the api_definition.json file.
# https://github.com/anyproto/anytype-heart/blob/main/core/api/docs/swagger.json

curl https://raw.githubusercontent.com/anyproto/anytype-heart/refs/heads/main/core/api/docs/openapi.json -o api_definition.json
if [ $? -ne 0 ]; then
    echo "Failed to fetch the API definition file."
    exit 1
fi