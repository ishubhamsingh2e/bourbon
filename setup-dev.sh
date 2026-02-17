#!/bin/bash

# Setup development environment for Bourbon

echo "🔧 Setting up Bourbon development environment..."

# 1. Configure git hooks
echo "deg configuring git hooks..."
git config core.hooksPath .githooks
chmod +x .githooks/pre-commit
echo "✅ Git hooks configured to use .githooks directory"

echo ""
echo "🎉 Development environment setup complete!"
