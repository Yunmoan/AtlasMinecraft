// Package world owns runtime state for one dimension. Region actors serialize
// chunk access; a worker pool loads or generates requested chunks and returns
// results through region mailboxes. World coordinates metadata, lighting,
// persistence and updates. It delegates save-file I/O to javasave, region-file
// storage to anvil, and protocol section/light encoding to chunkwire.
package world
