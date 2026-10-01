package world

// Stats is a best-effort snapshot. It never posts actor messages or waits for
// generation, so operators can inspect a busy world without joining its backlog.
type Stats struct {
	LoadedChunks      int64
	Regions           int
	MailboxMessages   int
	GenerationWorkers int
	ActiveGeneration  int64
	QueuedGeneration  int
}

func (w *World) Stats() Stats {
	s := Stats{LoadedChunks: w.loadedChunks.Load(), GenerationWorkers: w.workers, ActiveGeneration: w.activeJobs.Load()}
	w.mu.Lock()
	s.Regions = len(w.regions)
	for _, r := range w.regions {
		s.MailboxMessages += len(r.mailbox)
	}
	w.mu.Unlock()
	w.generation.mu.Lock()
	s.QueuedGeneration = len(w.generation.pending)
	w.generation.mu.Unlock()
	return s
}
