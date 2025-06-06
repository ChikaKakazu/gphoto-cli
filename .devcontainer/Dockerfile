FROM golang:1.24.4-bullseye

# Install additional tools for development
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js for Claude Code CLI
RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - && \
    apt-get install -y nodejs

# Install Claude Code CLI (if available via npm)
RUN npm install -g @anthropic-ai/claude-code || echo "Claude Code CLI not available via npm"

# Set up Go environment
ENV GOPATH=/go
ENV GOROOT=/usr/local/go
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

# Install common Go tools and Cobra generator
RUN go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install honnef.co/go/tools/cmd/staticcheck@latest && \
    go install github.com/spf13/cobra-cli@latest

# Create workspace directory
RUN mkdir -p /workspaces/gphoto-cli
WORKDIR /workspaces/gphoto-cli