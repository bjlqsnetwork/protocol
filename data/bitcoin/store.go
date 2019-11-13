/*

 */

package bitcoin

import (
	"fmt"
	"strings"

	"github.com/Oneledger/protocol/serialize"
	"github.com/Oneledger/protocol/storage"
	"github.com/pkg/errors"
)

var (
	ErrTrackerNotFound = errors.New("tracker not found")
)

type TrackerStore struct {
	State  *storage.State
	szlr   serialize.Serializer
	prefix []byte
}

func NewTrackerStore(prefix string, state *storage.State) *TrackerStore {
	return &TrackerStore{
		State:  state,
		szlr:   serialize.GetSerializer(serialize.PERSISTENT),
		prefix: storage.Prefix(prefix),
	}
}

// WithState updates the storage state of the tracker and returns the tracker address back
func (ts *TrackerStore) WithState(state *storage.State) *TrackerStore {
	ts.State = state
	return ts
}

func (ts *TrackerStore) Get(name string) (*Tracker, error) {

	key := keyFromName(name)

	key = append(ts.prefix, key...)
	exists := ts.State.Exists(key)
	if !exists {
		return nil, ErrTrackerNotFound
	}

	data, _ := ts.State.Get(key)

	d := &Tracker{}
	err := ts.szlr.Deserialize(data, d)
	if err != nil {
		return nil, errors.Wrap(err, "error de-serializing domain")
	}

	return d, nil

}

func (ts *TrackerStore) GetTrackerForLock() (*Tracker, error) {

	fmt.Println(" *************************************************************")
	start := append(ts.prefix, []byte("tracker_  ")...)
	end := append(ts.prefix, []byte("tracker_~~")...)

	var lowestAmount int64 = 999999999999999
	var tempTracker *Tracker = nil

	fmt.Println("111111111111111111111111111111111111111111111111111111111111111111111")
	doAscending := true
	ts.State.IterateRange(start, end, doAscending, func(k, v []byte) bool {

		fmt.Println("22222222222222222222222222222222222222222222222222222222222222222")
		d := &Tracker{}
		err := ts.szlr.Deserialize(v, d)
		if err != nil {
			fmt.Println("\n\n\n\n ERROR", err)
			return false
		}

		if d.IsAvailable() && d.GetBalance() < lowestAmount {
			tempTracker = d
			lowestAmount = d.CurrentBalance
		}

		// return false
		return false
	})

	if tempTracker == nil {
		return nil, errors.New("no tracker found")
	}

	return tempTracker, nil
}

func (ts *TrackerStore) SetTracker(name string, tracker *Tracker) error {

	tracker.Name = name

	key := keyFromName(name)
	key = append(ts.prefix, key...)

	data, err := ts.szlr.Serialize(tracker)
	if err != nil {
		return errors.Wrap(err, "error de-serializing domain")
	}

	return ts.State.Set(storage.StoreKey(key), data)
}

func (ts *TrackerStore) SetLockScript(lockAddress, lockScript []byte) error {
	key := append([]byte("lockscript:"), lockAddress...)

	return ts.State.Set(key, lockScript)
}

func (ts *TrackerStore) GetLockScript(lockAddress []byte) ([]byte, error) {
	key := append([]byte("lockscript:"), lockAddress...)

	return ts.State.Get(key)
}

func keyFromName(name string) []byte {
	return []byte(strings.ToLower(name))
}