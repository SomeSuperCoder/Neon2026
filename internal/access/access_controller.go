package access

import (
	"fmt"
	"sync"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/transaction"
)

// Package access implements the AccessController for validating file access permissions
// during instruction execution. It ensures that programs only access files they have
// declared with appropriate permissions.

// AccessController validates file access permissions during instruction execution
// This implements Requirements 4.1, 4.2, 4.3, 4.4, 4.5, 8.1, 8.2, 8.3, 8.4, 8.5
type AccessController struct {
	// accessLog tracks actual file accesses during execution
	accessLog map[filestore.FileID]transaction.AccessPermission
	// declaredAccess stores the declared access permissions from the instruction
	declaredAccess map[filestore.FileID]transaction.AccessPermission
	// mu protects concurrent access to the maps
	mu sync.RWMutex
}

// NewAccessController creates a new AccessController instance
func NewAccessController() *AccessController {
	return &AccessController{
		accessLog:      make(map[filestore.FileID]transaction.AccessPermission),
		declaredAccess: make(map[filestore.FileID]transaction.AccessPermission),
	}
}

// SetDeclaredAccess initializes the declared access permissions from an instruction's inputs
// This should be called before instruction execution begins
func (ac *AccessController) SetDeclaredAccess(inputs map[string]transaction.FileAccess) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Clear previous state
	ac.declaredAccess = make(map[filestore.FileID]transaction.AccessPermission)
	ac.accessLog = make(map[filestore.FileID]transaction.AccessPermission)

	// Populate declared access from instruction inputs
	for _, fileAccess := range inputs {
		// If a file is declared multiple times, use the highest permission level
		if existing, exists := ac.declaredAccess[fileAccess.FileID]; exists {
			if fileAccess.Permission > existing {
				ac.declaredAccess[fileAccess.FileID] = fileAccess.Permission
			}
		} else {
			ac.declaredAccess[fileAccess.FileID] = fileAccess.Permission
		}
	}
}

// ValidateAccess checks if a file access is allowed based on declared permissions
// This implements Requirements 4.5, 8.1, 8.2
func (ac *AccessController) ValidateAccess(fileID filestore.FileID, permission transaction.AccessPermission) error {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	// Check if file was declared in inputs (Requirement 4.5)
	declaredPerm, declared := ac.declaredAccess[fileID]
	if !declared {
		return fmt.Errorf("undeclared file access: file %s was not declared in instruction inputs", fileID.String())
	}

	// Check if requested permission matches or is less than declared permission
	// Write permission (2) requires Write to be declared
	// Read permission (1) can be satisfied by either Read or Write declaration
	if permission == transaction.Write {
		// Write access requires Write permission to be declared (Requirement 8.2)
		if declaredPerm != transaction.Write {
			return fmt.Errorf("permission violation: file %s declared as %s but write access attempted",
				fileID.String(), declaredPerm.String())
		}
	} else if permission == transaction.Read {
		// Read access is allowed if either Read or Write is declared
		if declaredPerm != transaction.Read && declaredPerm != transaction.Write {
			return fmt.Errorf("permission violation: file %s has invalid permission %d",
				fileID.String(), declaredPerm)
		}
	} else {
		return fmt.Errorf("invalid permission type: %d", permission)
	}

	return nil
}

// RecordAccess logs a file access during instruction execution
// This implements Requirement 8.4
func (ac *AccessController) RecordAccess(fileID filestore.FileID, permission transaction.AccessPermission) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Record the access, upgrading to Write if a previous Read was recorded
	if existing, exists := ac.accessLog[fileID]; exists {
		if permission > existing {
			ac.accessLog[fileID] = permission
		}
	} else {
		ac.accessLog[fileID] = permission
	}
}

// GetAccessLog retrieves the complete access log for the current execution
// This implements Requirement 8.4
func (ac *AccessController) GetAccessLog() map[filestore.FileID]transaction.AccessPermission {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	// Return a copy to prevent external modifications
	logCopy := make(map[filestore.FileID]transaction.AccessPermission, len(ac.accessLog))
	for fileID, perm := range ac.accessLog {
		logCopy[fileID] = perm
	}

	return logCopy
}

// ValidateAndRecord validates access permission and records it if valid
// This is a convenience method that combines ValidateAccess and RecordAccess
// This implements Requirements 8.1, 8.2, 8.3, 8.4, 8.5
func (ac *AccessController) ValidateAndRecord(fileID filestore.FileID, permission transaction.AccessPermission) error {
	// Validate first (Requirement 8.1, 8.2)
	if err := ac.ValidateAccess(fileID, permission); err != nil {
		// Halt execution on unauthorized access (Requirement 8.5)
		return err
	}

	// Record the access (Requirement 8.4)
	ac.RecordAccess(fileID, permission)

	return nil
}

// Reset clears all access logs and declared permissions
// This should be called between instruction executions
func (ac *AccessController) Reset() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.accessLog = make(map[filestore.FileID]transaction.AccessPermission)
	ac.declaredAccess = make(map[filestore.FileID]transaction.AccessPermission)
}

// GetDeclaredAccess returns a copy of the declared access permissions
func (ac *AccessController) GetDeclaredAccess() map[filestore.FileID]transaction.AccessPermission {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	// Return a copy to prevent external modifications
	declaredCopy := make(map[filestore.FileID]transaction.AccessPermission, len(ac.declaredAccess))
	for fileID, perm := range ac.declaredAccess {
		declaredCopy[fileID] = perm
	}

	return declaredCopy
}

// ValidateFinalAccess verifies that all actual accesses matched declared permissions
// This should be called after instruction execution completes
// This implements Requirement 8.1
func (ac *AccessController) ValidateFinalAccess() error {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	// Check that all accessed files were declared
	for fileID, actualPerm := range ac.accessLog {
		declaredPerm, declared := ac.declaredAccess[fileID]
		if !declared {
			return fmt.Errorf("undeclared file access detected: file %s was accessed but not declared", fileID.String())
		}

		// Verify that actual access didn't exceed declared permission
		if actualPerm == transaction.Write && declaredPerm != transaction.Write {
			return fmt.Errorf("permission violation detected: file %s was written but only declared as %s",
				fileID.String(), declaredPerm.String())
		}
	}

	return nil
}
