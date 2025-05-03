#!/bin/bash

# Build the frontend
cd web
npm install
npm run build

# Create dist directory in the project root if it doesn't exist
mkdir -p ../dist

# Copy the built frontend to the dist directory
cp -r dist/* ../dist/

# go back to root directory
cd ..

# Install Go dependencies
go build -o copy-paste

echo "Build completed! You can run the server with ./copy-paste"
