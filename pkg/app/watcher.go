package app

import (
	"time"

	fs "github.com/fsnotify/fsnotify"
)

// This is the amount of time to wait after changes before reacting to them.
// Debounce is done because moving files into the watched directories causes
// many rapid "Write" events to fire which would cause excessive Remove/Add
// method calls on the Library. To avoid this we accumulate the changes and
// only perform them once the events have stopped for this amount of time.
const debounceTimeout = time.Second * 5

// create, write, and chmod all require an add event
const addFlags = fs.Create | fs.Write | fs.Chmod

// remove, rename, write, and chmod all require a remove event
const removeFlags = fs.Remove | fs.Rename | fs.Write | fs.Chmod

// watch library paths and update Library with changes.
func startWatcher(a *App) {
	timer := time.NewTimer(debounceTimeout)
	addEvents := make(map[string]struct{})
	removeEvents := make(map[string]struct{})
	for {
		select {
		case e := <-a.Watcher.Events:
			if e.Op&removeFlags != 0 {
				removeEvents[e.Name] = struct{}{}
			}
			if e.Op&addFlags != 0 {
				addEvents[e.Name] = struct{}{}
			}
			// reset timer
			timer.Reset(debounceTimeout)
		case <-timer.C:
			eventCount := len(removeEvents) + len(addEvents)
			// handle remove events first
			if len(removeEvents) > 0 {
				for p := range removeEvents {
					a.Library.Remove(p)
				}
				// clear map
				removeEvents = make(map[string]struct{})
			}
			// then handle add events
			if len(addEvents) > 0 {
				for p := range addEvents {
					a.Library.Add(p)
				}
				// clear map
				addEvents = make(map[string]struct{})
			}
			if eventCount > 0 {
				buildFeed(a)
			}
			// reset timer
			timer.Reset(debounceTimeout)
		}
	}
}
