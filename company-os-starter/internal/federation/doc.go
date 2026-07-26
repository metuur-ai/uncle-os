// Package federation implements Option B multi-repo sync.
//
// Responsibility: manifest loading and validation, sparse-checkout, slice
// materialization at 0444/0555, and the lock. It depends on internal/workspace
// for path resolution and for MANIFEST_NAME/LOCK_NAME.
//
// Not implemented — Phase 6.
package federation
