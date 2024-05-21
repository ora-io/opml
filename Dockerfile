# How to run instructions:
# 1. Generate ssh command: ssh-keygen -t rsa -b 4096 -C "your_email@example.com"
#    - Save the key in local repo where Dockerfile is placed as id_rsa
#    - Add the public key to the GitHub account
# 2. Build docker image: docker build -t ubuntu-opml-dev .
# 3. Run the hardhat: docker run -it --rm --name ubuntu-opml-dev-container ubuntu-opml-dev bash -c "npx hardhat node"
# 4. Run the challange script on the same container: docker exec -it ubuntu-opml-dev-container bash -c "./demo/challenge_simple.sh"


# Use an official Ubuntu as a parent image
FROM ubuntu:22.04

# Set environment variables to non-interactive to avoid prompts during package installations
ENV DEBIAN_FRONTEND=noninteractive

# Update the package list and install dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    golang \
    wget \
    curl \
    python3 \
    python3-pip \
    python3-venv \
    unzip \
    file \
    openssh-client \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js and npm
RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash - && \
    apt-get install -y nodejs

# Copy SSH keys into the container
COPY id_rsa /root/.ssh/id_rsa
RUN chmod 600 /root/.ssh/id_rsa
# Configure SSH to skip host key verification
RUN echo "Host *\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config

# Set the working directory
WORKDIR /root

# Clone the OPML repository
RUN git clone git@github.com:ora-io/opml.git --recursive
WORKDIR /root/opml

# Build the OPML project
RUN make build

# Change permission for the challenge script
RUN chmod +x demo/challenge_simple.sh

# Default command
CMD ["bash"]
