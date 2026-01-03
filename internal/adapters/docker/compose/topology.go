package compose

import (
	"fmt"

	"github.com/simone-viozzi/bosun/internal/ports"
)

// topologicalSort sorts containers by dependency order using depth-first search.
// Dependencies come before dependents in the result (start order).
// Reverse the result for stop order (dependents before dependencies).
//
// Returns error if a dependency cycle is detected.
func topologicalSort(containers []ports.StackContainer) ([]ports.StackContainer, error) {
	if len(containers) == 0 {
		return nil, nil
	}

	// Build service name -> container map
	byService := make(map[string]*ports.StackContainer)
	for i := range containers {
		c := &containers[i]
		byService[c.ServiceName] = c
	}

	// Track visited and in-stack states for cycle detection
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	result := make([]ports.StackContainer, 0, len(containers))

	// DFS visit function
	var visit func(serviceName string) error
	visit = func(serviceName string) error {
		if visited[serviceName] {
			return nil // Already processed
		}

		if inStack[serviceName] {
			return fmt.Errorf("dependency cycle detected involving service: %s", serviceName)
		}

		container, exists := byService[serviceName]
		if !exists {
			// Dependency not in this stack - could be external or misconfigured
			// For M3, we skip it (best effort)
			return nil
		}

		inStack[serviceName] = true

		// Visit dependencies first (recursive)
		for _, dep := range container.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}

		inStack[serviceName] = false
		visited[serviceName] = true

		// Append after dependencies (topological order)
		result = append(result, *container)
		return nil
	}

	// Visit all containers
	for _, container := range containers {
		if err := visit(container.ServiceName); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// reverse returns a new slice with elements in reverse order.
func reverse(containers []ports.StackContainer) []ports.StackContainer {
	reversed := make([]ports.StackContainer, len(containers))
	for i, c := range containers {
		reversed[len(containers)-1-i] = c
	}
	return reversed
}
