package anvil

import (
	"errors"
	"os"
	"syscall"
	"time"
)

// Windows rejects replacing a file while a reader omits FILE_SHARE_DELETE.
// Our region locks exclude Atlas readers; antivirus and indexers can still
// hold short-lived handles. Keep the complete, synced temporary file and retry
// only those Windows errors, without ever deleting the current region first.
func replaceRegionFile(source, target string) error {
	deadline := time.Now().Add(time.Second)
	delay := 10 * time.Millisecond
	for {
		err := os.Rename(source, target)
		if err == nil || !transientReplacementError(err) {
			return err
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return err
		}
		time.Sleep(min(delay, remaining))
		delay = min(2*delay, 100*time.Millisecond)
	}
}

func transientReplacementError(err error) bool {
	return errors.Is(err, syscall.ERROR_ACCESS_DENIED) ||
		errors.Is(err, syscall.Errno(32)) || // ERROR_SHARING_VIOLATION
		errors.Is(err, syscall.Errno(33)) // ERROR_LOCK_VIOLATION
}
