// Package dukdakit provides an insanely easy game server framework.
// Build production-ready game servers in a snap!
//
// DukDak (뚝딱) means "in a snap" or "quickly" in Korean,
// representing our philosophy of making game server development
// ridiculously easy and fun.
//
// Features are organized into categories accessible via dot notation:
//   - dukdakit.Distributed.NewOptimistic() - Optimistic concurrency control
//   - dukdakit.Retry.New()                 - Retry mechanisms with circuit breaker
//   - dukdakit.Timex.DayElapsed()          - Time elapsed checking utilities
//   - More categories coming soon...
package dukdakit

// Version represents the current version of dukdakit
const Version = "v0.0.1"