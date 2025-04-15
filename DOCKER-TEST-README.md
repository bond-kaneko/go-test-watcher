# Docker fsnotify Test

This directory contains tests to verify the behavior of fsnotify in different Docker environments.

## Prerequisites

- Docker
- Docker Compose

## Test Structure

The tests are designed to check fsnotify functionality in three different scenarios:

1. **Basic container test**: Files are created/modified/deleted within the container
2. **Bind mount test**: Tests fsnotify on a directory mounted from the host
3. **Volume test**: Tests fsnotify on a Docker volume

## Setup

1. Create required directories:

```sh
mkdir -p logs test-mount
```

2. Make sure the project dependencies are in your go.mod file:

```sh
go get github.com/fsnotify/fsnotify
```

## Running the Tests

You can run the tests using the provided docker-compose file:

```sh
# Run all tests
docker-compose -f docker-compose.test.yml up

# Run a specific test
docker-compose -f docker-compose.test.yml up basic-test
docker-compose -f docker-compose.test.yml up bind-mount-test
docker-compose -f docker-compose.test.yml up volume-test
```

## Test Interaction

For the bind-mount test, you can create, modify and delete files in the `test-mount` directory to see how fsnotify responds to external changes.

Example:

```sh
# Create a file
echo "Hello World" > test-mount/test-external.txt

# Modify the file
echo "Updated content" >> test-mount/test-external.txt

# Delete the file
rm test-mount/test-external.txt
```

## Expected Results

- **Basic test**: Should detect all file events properly
- **Bind mount test**: Should detect file events from both inside the container and from the host
- **Volume test**: Should detect all file events within the container, similar to the basic test

## Troubleshooting

1. If events are not being detected on bind mounts, it could be due to issues with the host filesystem or Docker setup.

2. On macOS, Docker Desktop uses a VM which can affect inotify events. You might experience delays or missed events.

3. On Windows, similar limitations can occur with WSL2.

## Notes

- The test program will automatically create, modify, and delete test files in the watched directory.
- Output is shown in real-time in the Docker logs.
- Press Ctrl+C to stop the tests. 
