# Grace

Grace is a Go package designed to assist in implementing graceful shutdown processes for Go applications. It provides a straightforward and efficient way to manage long-running processes, background tasks, or any operations that need to be cleanly shut down when your application receives a termination signal such as SIGINT or SIGTERM.

## Features

- **Graceful Shutdown**: Grace offers a simple yet powerful mechanism to handle graceful shutdowns, ensuring all your operations are properly completed before the application exits.
- **Concurrency Safe**: With built-in concurrency control, Grace safely manages multiple operations across different goroutines, ensuring thread-safety.
- **Singleton Design**: Grace implements a singleton design pattern, ensuring that only one instance of the shutdown manager is created throughout the application lifecycle. This approach simplifies managing graceful shutdowns and concurrency across your entire application by providing a central point of control.

## Installation

To use Grace in your Go project, run `go get github.com/TechMDW/grace` in your terminal.

## Usage

Import the Grace package into your Go application:

```go
import "github.com/TechMDW/grace"
```

Below is a simple example demonstrating how to use Grace to handle a graceful shutdown:

```go
package main

import (
  "context"
  "os"
  "os/signal"
  "syscall"
  "time"

  "github.com/TechMDW/grace"
)


func main() {
	// Some program logic
	go func() {
		grace.Add()          // Register a new task
		grace.AddX(2)        // Register multiple tasks at once
		defer grace.Done()   // Unregister a task
		defer grace.DoneX(2) // Unregister multiple tasks at once
		time.Sleep(10 * time.Second)
	}()

	// Setup a signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a termination signal
	// Note: This will block until a SIGINT or SIGTERM is received
	<-sigChan

	// Add context with timeout so we don't wait forever just in case
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Waiting for all resources to finish:", grace.Count())
	// This will block until all resources are finished or the context timeouts/cancel.
	grace.WaitCtx(ctx)
	os.Exit(0)
}
```

This example sets up a basic signal handler that waits for SIGINT or SIGTERM. Upon receiving one of these signals, it initiates the graceful shutdown process with a 30-second timeout.

## License

This project is licensed under the [Apache License 2.0](LICENSE).
