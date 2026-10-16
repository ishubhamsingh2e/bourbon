#!/bin/bash

# Bourbon Framework Installer
# This script installs the Bourbon CLI globally

set -e

echo "🥃 Installing Bourbon Framework CLI..."

# Install the CLI
echo "📦 Installing bourbon command..."
go install ./cmd/bourbon

# Check if installed
if [ -f "$HOME/go/bin/bourbon" ]; then
    echo "✅ Bourbon CLI installed successfully!"
else
    echo "❌ Installation failed. Please check your Go installation."
    exit 1
fi

# Check if PATH already contains ~/go/bin
if [[ ":$PATH:" == *":$HOME/go/bin:"* ]]; then
    echo "✅ ~/go/bin is already in your PATH"
else
    echo "⚠️  Adding ~/go/bin to your PATH..."
    
    # Detect shell
    SHELL_NAME=$(basename "$SHELL")
    
    case "$SHELL_NAME" in
        zsh)
            echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
            echo "✅ Added to ~/.zshrc"
            echo "📝 Run: source ~/.zshrc"
            ;;
        bash)
            echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc
            echo "✅ Added to ~/.bashrc"
            echo "📝 Run: source ~/.bashrc"
            ;;
        fish)
            echo "fish_add_path ~/go/bin" >> ~/.config/fish/config.fish
            echo "✅ Added to Fish config"
            ;;
        *)
            echo "⚠️  Unknown shell: $SHELL_NAME"
            echo "Please add this to your shell config manually:"
            echo 'export PATH="$HOME/go/bin:$PATH"'
            ;;
    esac
fi

echo ""
echo "🎉 Installation complete!"
echo ""
echo "To use bourbon immediately in this terminal:"
echo "  export PATH=\"\$HOME/go/bin:\$PATH\""
echo ""
echo "Or restart your terminal, then try:"
echo "  bourbon version"
echo "  bourbon new my-project"
echo ""
echo "📚 Documentation:"
echo "  README.md - Framework overview"
echo "  docs/ - Comprehensive architecture and usage guides"
echo ""
echo "Happy coding with Bourbon! 🥃"
