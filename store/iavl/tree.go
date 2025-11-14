package iavl

import (
	"fmt"

	"github.com/cosmos/iavl"
	idb "github.com/cosmos/iavl/db"
)

var (
	_ Tree = (*immutableTree)(nil)
	_ Tree = (*mutableTreeWrapper)(nil)
)

// mutableTreeWrapper wraps iavl.MutableTree to implement the Tree interface
type mutableTreeWrapper struct {
	*iavl.MutableTree
}

// DeleteVersionsTo implements Tree interface for mutableTreeWrapper
// Note: moca-iavl may not have this method, so we return an error
func (mtw *mutableTreeWrapper) DeleteVersionsTo(version int64) error {
	// moca-iavl doesn't have DeleteVersionsTo method
	// Return an error indicating it's not implemented
	return fmt.Errorf("DeleteVersionsTo is not implemented in moca-iavl")
}

// TraverseStateChanges implements Tree interface for mutableTreeWrapper
func (mtw *mutableTreeWrapper) TraverseStateChanges(startVersion, endVersion int64, fn func(version int64, changeSet interface{}) error) error {
	// moca-iavl may not have this method
	return fmt.Errorf("TraverseStateChanges is not implemented in moca-iavl")
}

// Hash implements Tree interface for mutableTreeWrapper
// moca-iavl's Hash() returns ([]byte, error), but Tree interface requires []byte
func (mtw *mutableTreeWrapper) Hash() []byte {
	hash, err := mtw.MutableTree.Hash()
	if err != nil {
		panic(fmt.Errorf("failed to get hash: %w", err))
	}
	return hash
}

// WorkingHash implements Tree interface for mutableTreeWrapper
// moca-iavl's WorkingHash() returns ([]byte, error), but Tree interface requires []byte
func (mtw *mutableTreeWrapper) WorkingHash() []byte {
	hash, err := mtw.MutableTree.WorkingHash()
	if err != nil {
		panic(fmt.Errorf("failed to get working hash: %w", err))
	}
	return hash
}

// LoadVersionForOverwriting implements Tree interface for mutableTreeWrapper
// moca-iavl's LoadVersionForOverwriting() returns (int64, error), but Tree interface requires error
func (mtw *mutableTreeWrapper) LoadVersionForOverwriting(targetVersion int64) error {
	_, err := mtw.MutableTree.LoadVersionForOverwriting(targetVersion)
	return err
}

type (
	// Tree defines an interface that both mutable and immutable IAVL trees
	// must implement. For mutable IAVL trees, the interface is directly
	// implemented by an iavl.MutableTree. For an immutable IAVL tree, a wrapper
	// must be made.
	Tree interface {
		Has(key []byte) (bool, error)
		Get(key []byte) ([]byte, error)
		Set(key, value []byte) (bool, error)
		Remove(key []byte) ([]byte, bool, error)
		SaveVersion() ([]byte, int64, error)
		Version() int64
		Hash() []byte
		WorkingHash() []byte
		VersionExists(version int64) bool
		DeleteVersionsTo(version int64) error
		GetVersioned(key []byte, version int64) ([]byte, error)
		GetImmutable(version int64) (*iavl.ImmutableTree, error)
		SetInitialVersion(version uint64)
		Iterator(start, end []byte, ascending bool) (idb.Iterator, error)
		AvailableVersions() []int
		LoadVersionForOverwriting(targetVersion int64) error
		TraverseStateChanges(startVersion, endVersion int64, fn func(version int64, changeSet interface{}) error) error
	}

	// immutableTree is a simple wrapper around a reference to an iavl.ImmutableTree
	// that implements the Tree interface. It should only be used for querying
	// and iteration, specifically at previous heights.
	immutableTree struct {
		*iavl.ImmutableTree
	}
)

func (it *immutableTree) Set(_, _ []byte) (bool, error) {
	panic("cannot call 'Set' on an immutable IAVL tree")
}

func (it *immutableTree) Remove(_ []byte) ([]byte, bool, error) {
	panic("cannot call 'Remove' on an immutable IAVL tree")
}

func (it *immutableTree) SaveVersion() ([]byte, int64, error) {
	panic("cannot call 'SaveVersion' on an immutable IAVL tree")
}

func (it *immutableTree) DeleteVersionsTo(_ int64) error {
	panic("cannot call 'DeleteVersionsTo' on an immutable IAVL tree")
}

func (it *immutableTree) SetInitialVersion(_ uint64) {
	panic("cannot call 'SetInitialVersion' on an immutable IAVL tree")
}

func (it *immutableTree) VersionExists(version int64) bool {
	return it.Version() == version
}

func (it *immutableTree) GetVersioned(key []byte, version int64) ([]byte, error) {
	if it.Version() != version {
		return nil, fmt.Errorf("version mismatch on immutable IAVL tree; got: %d, expected: %d", version, it.Version())
	}

	return it.Get(key)
}

func (it *immutableTree) GetImmutable(version int64) (*iavl.ImmutableTree, error) {
	if it.Version() != version {
		return nil, fmt.Errorf("version mismatch on immutable IAVL tree; got: %d, expected: %d", version, it.Version())
	}

	return it.ImmutableTree, nil
}

func (it *immutableTree) AvailableVersions() []int {
	return []int{}
}

func (it *immutableTree) LoadVersionForOverwriting(targetVersion int64) error {
	panic("cannot call 'LoadVersionForOverwriting' on an immutable IAVL tree")
}

func (it *immutableTree) Hash() []byte {
	hash, err := it.ImmutableTree.Hash()
	if err != nil {
		panic(fmt.Errorf("failed to get hash: %w", err))
	}
	return hash
}

func (it *immutableTree) WorkingHash() []byte {
	panic("cannot call 'WorkingHash' on an immutable IAVL tree")
}

func (it *immutableTree) TraverseStateChanges(startVersion, endVersion int64, fn func(version int64, changeSet interface{}) error) error {
	panic("cannot call 'TraverseStateChanges' on an immutable IAVL tree")
}
