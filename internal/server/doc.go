// Package server implements the Java 1.21.1 connection lifecycle and gameplay
// request handling. It coordinates players, commands, visibility and dimensions;
// authoritative chunk operations are delegated to world. Connection readers and
// per-player writers run independently, with bounded outbound queues.
package server
